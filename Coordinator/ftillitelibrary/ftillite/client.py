# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from __future__ import annotations
from codecs import ignore_errors # Note: needs to be at the top of the file
# Demo for the ftillite Python frontend
# =====================================
#
# Revision history:
#
##################################################################

from ftillite.compute_manager import ComputeManager

import re
import pickle
import os
import logging
import weakref

# Type hint libs
from typing import Tuple, Union

class FTILConf:
    def __init__(self):
        self.app_name = ""
        self.rabbitmq_conf = None
        self.client_logger = None
        self.compute_manager_logger = None
        self.segment_client_logger = None
        self.peer_names = []
  
    def set_app_name(self, name):
        self.app_name = name
        return self

    def set_rabbitmq_conf(self, rmq_conf):
        self.rabbitmq_conf = rmq_conf
        return self

    def set_all_loggers(self, logger):
        self.client_logger = logger
        self.compute_manager_logger = logger
        self.segment_client_logger = logger
        return self

    def set_client_logger(self, client_logger):
        self.client_logger = client_logger
        return self

    def set_compute_manager_logger(self, compute_manager_logger):
        self.compute_manager_logger = compute_manager_logger
        return self

    def set_segment_client_logger(self, segment_client_logger):
        self.segment_client_logger = segment_client_logger
        return self

    def set_peer_names(self, peer_names):
        self.peer_names = peer_names
        return self

def close():
    # Closes the FTIL Context and backend object
    backend.clear_variable_registry()
    backend.clear_peer_var_store()

class FTILContext:

    def __init__(self, conf):
        self._backend = ComputeManager(conf.rabbitmq_conf, conf.compute_manager_logger, conf.segment_client_logger, conf.peer_names )
        global backend
        backend = self._backend # Sets the global backend object
        self.rabbitmq_conf = conf.rabbitmq_conf
        rc = self._backend.init(self)
        self.CoordinatorID = Node(self, rc[0][0], rc[0][1])
        self.Nodes = {rc[0][0]: self.CoordinatorID}
        for r in rc[1:]:
            self.Nodes[r[0]] = Node(self, r[0], r[1])
        self._scope_stack = [NodeSet(self.Nodes.values())]
        self.MyID = IntArrayIdentifier(fc=self, handle="0")
        self.logger = conf.client_logger
        self.dead_handles = {}

        # Handle when the context is closed - clears all remaining variables
        weakref.finalize(self, close)
        
    def queue_for_deletion_append(self, handle: str, scope: list[int]):
        # Adds / Marks variables for deletion, bypassing the Garbage Collector
        for s in scope:
            if s not in self.dead_handles:
                self.dead_handles[s] = []
            self.dead_handles[s].append(handle)
        
    def queue_for_deletion_clear(self):
        # Clears the variable registry of variables marked for deletion
        # for h in self.dead_handles:
        self._backend.del_handles(self.dead_handles)
        self.dead_handles = {} 
        
    def clear(self):
        # Triggers the variable stores to be cleared on all peer nodes
        self._backend.clear_peer_var_store()  

    def ping(self):
        res = {}
        for s in self._backend.segment_clients:
            s.connect()
            s.channel = s.connection.channel()
            try:
                s.channel.queue_declare(queue=s.in_queue, passive=True)
                res[s.name] = True
            except Exception as ex:
                self._print(f"Unable to connect to segment manager {s.num} - {s.name} with exception {ex}", logging.ERROR)
                res[s.name] = False
        return res

    def _print(self, value, loglevel=logging.INFO):
        if self.logger:
            self.logger.log(loglevel, f"CLIENT - {value}")

    def _extract_value(self, rc):
        if rc[0] == "array":
            if rc[1] == "i":
                return IntArrayIdentifier(self, rc[2])
            elif rc[1] == "f":
                return FloatArrayIdentifier(self, rc[2])
            elif rc[1] == "I":
                return Ed25519IntArrayIdentifier(self, rc[2])
            elif rc[1] == "E":
                return Ed25519ArrayIdentifier(self, rc[2])
            elif rc[1][0] == "b":
                return BytearrayArrayIdentifier(self, rc[2], rc[1])
            else:
                raise ValueError("Unknown typecode.")
        elif rc[0] == "listmap":
            return ListMapIdentifier(self, rc[1], rc[2])
        else:
            raise ValueError("Unknown type.")

    def _exec_command(self, command):
        scope = [x.num() for x in self.scope()]
        self._print(f"COMMAND = {scope}: {command}")
        rc = self._backend.process_command(command, scope)
        self._print(f"RC = {rc}")
        # "process_command" triggers an MQ request with its parameter.
        # rc is the string that is the MQ response.
        # Typically, it comes in batches of (type, typecode, handle)
        return self._parse_result(rc)

    def _parse_result(self, rc):
        if len(rc) == 0:
            return
        rc = rc.split(" ")
        if rc[0] == "intlist":
            return [int(x) for x in rc[1:] if x != '']
        elif rc[0] == "floatlist":
            return [float(x) for x in rc[1:] if x != '']
        elif rc[0] == "bytearraylist":
            byte_array_list = []
            byte_array = []
            for x in rc[1:]:
                if ',' not in x:
                    byte_array.append(x)
                else:
                    s = x.split(',')
                    byte_array.append(s[0])
                    byte_array_list.append(byte_array)
                    byte_array = []
                    byte_array.append(s[1])
            byte_array_list.append(byte_array)
            return byte_array_list
        elif rc[0] == "int":
            return int(rc[1])
        elif rc[0] == "bool":
            return int(rc[1]) != 0
        elif rc[0] == "map":
            retval = {}
            del rc[0]
            while len(rc) != 0:
                retval[self.Nodes[int(rc[0])]] = self._extract_value(rc[1:4])
                retval[self.Nodes[int(rc[0])]]._scope = NodeSet({ self.Nodes[x] for x in self._backend.variable_registry[rc[3]][2]})
                del rc[:4]
            return retval
        elif rc[0] == "error":
            self.logger.ERROR(f"Error thrown: {rc}")
            return
        else:
            retval = []
            # rc is type+typecode+handle (only for new handles).
            while len(rc) != 0:
                retval.append(self._extract_value(rc[:3]))
                del rc[:3]
            if len(retval) > 1:
                return retval
            else:
                return retval[0]

    def _save_handle(self, handle):
        return self._backend.save_handle(handle)

    def _load_handle(self, handle, registry_entry, typecode):
        return self._backend.load_handle(handle, registry_entry, typecode)

    def del_handle(self, handle):
        self._backend.del_handle(handle)

    def transmit(self, opcode, typecode, arguments):
        return self._parse_result(
            self._backend.transmit(opcode, typecode, arguments))

    def __setattr_(self, name, value):
        raise RuntimeError("Attempting to modify a constant value.")

    def __delattr_(self, name):
        raise RuntimeError("Attempting to delete a protected attribute.")

    def _push_scope(self, new_scope):
        # Not doing much sanity checking here. Input is supposed to be legal.
        if new_scope.issubset(self._scope_stack[-1]) and len(new_scope) != 0 and isinstance(new_scope, NodeSet):
            self._scope_stack.append(new_scope)
        else:
            raise ValueError("Invalid new scope.")

    def _pop_scope(self):
        if len(self._scope_stack) > 1:
            del self._scope_stack[-1]
        else:
            raise BufferError("Scope stack already at top.")

    def scope(self):
        return self._scope_stack[-1].copy()

    # log_message writes the specified message into the log of each node
    def log_message(self, message):
        return self._exec_command("log_message 0 " + message)

    def log_stats(self, clear_commands = False):
        c = "log_stats 0"
        if clear_commands:
            c += " 1"
        return self._exec_command(c)

    def singleton(self, typecode, value):
        if typecode == 'i' and type(value) is int:
            return self._exec_command("newilist 1 " + str(value))
        if typecode == 'f' and type(value) is float:
            return self._exec_command("newflist 1 " + str(value))
        else:
            raise TypeError("Not a valid type for singleton creation.")

    def array(self, typecode, length=0, value=None):
        # Note length will be the equivalent of value if it's a list
        # See Array constructor documentation
        if type(length) is list:
            if typecode == 'i':
                for x in length:
                    if type(x) != int:
                        raise TypeError("Python list must contain integer types only")
                return self._exec_command("newilist 1 " +
                                          (" ".join(str(x) for x in length)))
            elif typecode == 'f':
                return self._exec_command("newflist 1 " +
                                          (" ".join(str(float(x)) for x in length)))
            else:
                raise TypeError(
                    "Only int and float lists can construct arrays.")
        else:
            length = self.promote(length, "i")
            if value == None:
                return self._exec_command("newarray 1 " + typecode + " "
                                          + length[0].handle())
            else:
                value = self.promote(value, typecode)
                return self._exec_command("newarray 1 " + typecode + " "
                                          + length[0].handle() + " " + value[0].handle())

    def listmap(self, *args, **kwargs):

        keys = []
        listmap_len = None
        typecodes = None
        arrays = None
        arg_len = len(args)

        if arg_len == 0:
            raise ValueError("listmap requires at least one argument")

        # fc.listmap(typecodes, arrays, order = "any")
        if arg_len >= 2 and type(args[0]) is str and isinstance(args[1], list):
            typecodes = typecode_list(args[0])
            arrays = args[1]
            order = None if arg_len == 2 else args[2]

            if len(arrays) != len(typecodes):
                raise ValueError("number of arrays items does not match number of typecodes provided")
        # fc.listmap(typecodes)
        elif arg_len <= 2 and type(args[0]) is str: 
            typecodes = typecode_list(args[0])
            order = None if arg_len == 1 else args[1]
        # fc.listmap(arrays, order = "any")
        elif arg_len <= 2 and isinstance(args[0], list):
            arrays = args[0]
            order = None if arg_len == 1 else args[1]
        else:
            raise ValueError("invalid arguments for listmap constructor")

        if not order:
            if "order" in kwargs:
                order = kwargs['order']
            else:
                order="any"
        # check order is set to a valid value
        if order not in ["any", "pos", "rnd"]:
            raise ValueError(f"invalid order value: {order}, must be either any, pos or rnd")

        if arrays:
            # Promote and Cast Type
            for i, a in enumerate(arrays):
                copied = False
                if typecodes:
                    arrays[i], copied = self.promote(a, typecodes[i])
                else:
                    arrays[i], copied = self.promote(a)

                if not copied:
                    arrays[i] = arrays[i].copy()

            # Perform Broadcast
            listmap_len = self.calc_broadcast_length(arrays)
            for key in arrays:
                if not _equal_int(key.len(), listmap_len):
                    key = key.broadcast_value(listmap_len)[0]
                if not isinstance(key, ArrayIdentifier):
                    raise ValueError(f"arrays param must contain list of ArrayIdentifier only, not {type(key)}")
                keys.append(key)
        else:
            for tc in typecodes:
                keys.append(self.array(tc, 0))

        if len(keys) == 0:
            raise ValueError("no keys provided")

        if order == "pos" and not self.keys_unique(keys):
            raise ValueError("keys are not unique for order 'pos'")
        return self._exec_command(f"newlistmap 1 {''.join([k.typecode() for k in keys])} {order} {' '.join([k.handle() for k in keys])}")
    
    def keys_unique(self, keys):
        if not isinstance(keys, list):
            raise ValueError("keys is not of type list")
        for i, k in enumerate(keys):
            keys[i] = self.promote(k)[0]
        return self._exec_command(f"listmap_keys_unique 0 {' '.join([k.handle() for k in keys])}")

    def save(self, obj, destination):
        # This method is not reentrant.
        self._backend.start_remote_save(destination)
        file = open(f"savecompute_{destination}.pickle", "wb")
        pickle.dump(obj, file)
        file.close()
        del file
        self._backend.finish_remote_save()

    def load(self, destination):
        try:
            # This method is not reentrant.
            self._backend.start_remote_load(destination)
            file = open(f"savecompute_{destination}.pickle", "rb")
            obj = pickle.load(file)
            file.close()
            del file
            self._backend.finish_remote_load()
        except Exception as err:
            if 'obj' in locals():
                del obj
            raise
        return obj

    def delete(self, destination):
        self._exec_command(f"delfile 0 {destination}")
        os.remove(f"savecompute_{destination}.pickle")

    def auxdb_read(self, query, typecodes=None):
        if not typecodes:
            return self._exec_command(f'auxdb_read 0 __{query}__')

        typecodes = typecode_list(typecodes)
        res = self._exec_command(f'auxdb_read {len(typecodes)} {" ".join(typecodes)} __{query}__')
            
        if len(typecodes) == 1:
            return res
        else:
            return list(res)

    def auxdb_write(self, table_name, col_names, data):
        if not isinstance(col_names, list):
            raise ValueError("col_names must be a python list")
        
        if not isinstance(data, list):
            raise ValueError("data must be a python list")
        for d in data:
            if not isinstance(d, ArrayIdentifier):
                raise ValueError("all values in data must be an ArrayIdentifier")

        if len(col_names) != len(data):
            raise ValueError("col_names and data must have the same number of elements")

        # TODO: Check all arrays are of the same length

        return self._exec_command(f'auxdb_write 0 {table_name} {" ".join(col_names)} {" ".join([d.handle() for d in data])}')
        
    def randomarray(self, typecode, length, *args, **kwargs):
        length = self.promote(length, "i")[0]

        if typecode == "i" or typecode == "f":
            if typecode == "i":
                min_val = 0 if len(args) < 2 else args[0]
                max_val = 1 if len(args) < 2 else args[1]
            elif typecode == "f":
                min_val = 0.0 if len(args) < 2 else args[0]
                max_val = 1.0 if len(args) < 2 else args[1]
            if min_val >= max_val:
                raise ValueError("minimum value cannot equal or exceed maximum value.")
            
            min_val = self.promote(min_val, typecode)[0]
            max_val = self.promote(max_val, typecode)[0]

            return self._exec_command(f'randomarray 1 {typecode} {length.handle()} {min_val.handle()} {max_val.handle()}')
        elif typecode == "I":
            nonzero = "True" if len(args) < 1 or args[0] else "False"

            return self._exec_command(f'randomarray 1 I {length.handle()} {nonzero}')
        elif typecode[0] == "b":
            return self._exec_command(f'randomarray 1 {typecode} {length.handle()}')
        else:
            raise TypeError("Only int, float or ed25519Int random arrays are supported.")

    def randomperm(self, length, n=None):
        length =  self.promote(length, "i")[0]
        if n is None:
            n = length
        else:
            n = self.promote(n, "i")[0]
        return self._exec_command(f'randomperm 1 {length.handle()} {n.handle()}')

    def aes256_keygen(self):
        length = self.promote(1, "i")[0]
        
        return self._exec_command(f'randomarray 1 b32 {length.handle()}')

    def grain128aeadv2_keygen(self):
        length = self.promote(1, "i")[0]
        
        return self._exec_command(f'randomarray 1 b16 {length.handle()}')

    def ecdsa256_keygen(self):
        privateKey = self._exec_command(f'ecdsa256_keygen 1')
        publicKey = self._exec_command(f'ecdsa256_public_key 1 {privateKey.handle()}')

        return (privateKey, publicKey)

    def rsa3072_keygen(self):
        privateKey = self._exec_command(f'rsa3072_keygen 1')
        publicKey = self._exec_command(f'rsa3072_public_key 1 {privateKey.handle()}')

        return (privateKey, publicKey)

    def calc_broadcast_length(self, params):
        if isinstance(params, list):
            for i, p in enumerate(params):
                params[i] = self.promote(p)[0]
        else:
            raise ValueError("params must be a python list")
        param_lens = [p.len() for p in params]
        handles = [l.handle() for l in param_lens]
        return self._exec_command(f"calc_broadcast_length 1 {' '.join(handles)}")

    def verify_context(self, params):
        for e in params:
            if isinstance(e, ManagedObject) and e.context() != self:
                raise ValueError("ManagedObject belongs to a foreign context.")

    def promote(self, obj, typecode=None):
        is_copy = False
        if type(obj) is int:
            obj = self.singleton("i", obj)
            is_copy = True
        elif type(obj) is float:
            obj = self.singleton("f", obj)
            is_copy = True
        elif isinstance(obj, list) and len(obj) > 0:
            if isinstance(obj[0], int):
                obj = self.array('i', obj)
            elif isinstance(obj[0], float):
                obj = self.array('f', obj)
            else:
                raise ValueError(f"cannot promote python list of type {type(obj[0])}.")
        if typecode != None:
            (obj, is_copy2) = obj.promote(typecode)
            is_copy = is_copy or is_copy2
        return (obj, is_copy)

    def branch2node(self, branches):
        raise NotImplementedError("Awaiting Golang implementation.")

    def arange(self, length):
        length = self.promote(length, 'i')[0]
        return self._exec_command("arange 1 "+length.handle())


class ManagedObject:
    def __init__(self, fc):
        self._context = fc

    def context(self):
        return self._context


class Node(ManagedObject):
    def __init__(self, fc, num, name):  # May include other properties, too.
        super().__init__(fc)
        self._num = num
        self._name = name

    def num(self):
        return self._num

    def name(self):
        return self._name
    # During standard use, Nodes do not need __hash__ and __eq__ because they
    # are singletons. However, to survive a save/load operation, we need to be
    # able to match loaded Nodes against existing ones.
    # Here, we determine equality solely based on the ".num()" property, which is
    # somewhat simplistic.

    def __hash__(self):
        return self.num()

    def __eq__(self, other):
        return self.num() == other.num()


class NodeSet(set, ManagedObject):
    def __init__(self, val):
        if type(val) is Node:
            set.__init__(self, [val])
            ManagedObject.__init__(self, val.context())
        elif isinstance(val, FTILContext):
            set.__init__(self)
            ManagedObject.__init__(self, val)
        else:
            if isinstance(val, NodeSet):
                fc = val.context()
            else:
                fc = get_context(val)
            for e in val:
                if type(e) is not Node:
                    raise TypeError("Elements sent to NodeSet are not nodes.")
            set.__init__(self, val)
            ManagedObject.__init__(self, fc)

    def __ior__(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        set.__ior__(self, other)
        self = NodeSet(self)

    def __ixor__(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        set.__ixor__(self, other)
        self = NodeSet(self)

    def __or__(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        return NodeSet(set.__or__(self, other))

    def __ror__(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        return NodeSet(set.__ror__(self, other))

    def __rxor__(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        return NodeSet(set.__rxor__(self, other))

    def __xor__(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        return NodeSet(set.__xor__(self, other))

    def add(self, other):
        if type(other) is not Node:
            raise TypeError("Unexpected parameter type. Expected Node.")
        if self.context() != other.context():
            raise ValueError("NodeSet and parameter contexts mismatch.")
        set.add(self, other)
        self = NodeSet(self)

    def difference(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        return NodeSet(set.difference(self, other))

    def symmetric_difference(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        return NodeSet(set.symmetric_difference(self, other))

    def symmetric_difference_update(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        set.symmetric_difference_update(self, other)
        self = NodeSet(self)

    def union(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        return NodeSet(set.union(self, other))

    def update(self, other):
        if isinstance(other, FTILContext):
            raise TypeError("Unexpected parameter type: FTILContext.")
        other = NodeSet(other)
        if self.context() != other.context():
            raise ValueError("NodeSets contexts mismatch.")
        set.update(self, other)
        self = NodeSet(self)

    def copy(self):
        return NodeSet(set.copy(self))


class on:
    def __init__(self, new_scope):
        self.new_scope = NodeSet(new_scope)
        if len(self.new_scope) == 0:
            raise ValueError("Scope cannot be empty.")
        self.context = self.new_scope.context()
        if not self.new_scope.issubset(self.context.scope()):
            raise ValueError("New scope must be a subset of existing scope.")

    def __enter__(self):
        self.context._push_scope(self.new_scope)
        return self.context.scope()

    def __exit__(self, exc_type, exc_value, traceback):
        self.context._pop_scope()


class Identifier(ManagedObject):
    def __init__(self, fc, **kwargs):
        super().__init__(fc)
        if not isinstance(self.context().scope(), NodeSet):
            raise ValueError("scope is not of type NodeSet")
        self._scope = self.context().scope()

    def scope(self):
        if not isinstance(self._scope, NodeSet):
            raise ValueError("scope is not of type NodeSet")
        return self._scope

    def width(self):
        return len(typecode_list(self.typecode()))

    def auxdb_read(self, query):
        flat = self.context().auxdb_read(query, self.typecode())
        if type(flat) is list:
            self.unflatten(flat)
        else:
            self.unflatten([flat])

    def typecode(self):
        return "".join([x.typecode() for x in self.flatten()])

    def __getstate__(self):
        state = {k: v for k, v in self.__dict__.items() if k != '_context'}
        state["_scope"] = [x.num() for x in self._scope]
        return state

    def __setstate__(self, state):
        # Python requires us to have a global "backend" object for this.
        self._context = backend.context()
        self.__dict__.update(state)
        if self._context == None:
            raise RuntimeError("Cannot load when not in a session.")
        self._scope = NodeSet([self.context().Nodes[x] for x in state["_scope"]])

    # log_value writes the value of the current variable into the log of each node.
    def log_value(self, label = None):
        if label is None:
            label = self.handle()
        # preserve spaces so they aren't lost in split to args on GoLang side.
        label = label.replace(" ", "~")

        self._context._exec_command(f'log_variable 0 {label} {self.handle()}')

class ArrayIdentifier(Identifier):
    def _handle_native_type(self, v):
        if type(v) == float or (isinstance(v, list) and all(isinstance(x, float) for x in v)):
            v = self.context().promote(v, 'f')[0]
        elif type(v) == int or (isinstance(v, list) and all(isinstance(x, int) for x in v)):
            v = self.context().promote(v, 'i')[0]
        return v

    def _operator(self, op, lhs, rhs, handles_returned=1, type_checker=None):

        # Process native values        
        lhs = self._handle_native_type(lhs)
        rhs = self._handle_native_type(rhs)
        tc_l = None
        tc_r = None

        if type_checker is not None and type_checker(lhs, rhs):
            pass
        elif type(lhs) == type(rhs):
            tc_l = self.typecode()
            tc_r = self.typecode()
        elif type(lhs) == BytearrayArrayIdentifier and type(rhs) in [int, IntArrayIdentifier]:
            tc_l = lhs.typecode()
            tc_r = 'i'
        elif isinstance(lhs, IntArrayIdentifier):
            lhs = lhs.astype(rhs.typecode())
        elif isinstance(rhs, IntArrayIdentifier):
            rhs = rhs.astype(lhs.typecode())
        else:
            raise TypeError(
            "Operands are of not compatible types")

        if tc_l is not None: 
            lhs = self.context().promote(lhs, tc_l)[0]
        if tc_r is not None:
            rhs = self.context().promote(rhs, tc_r)[0]

        lhslength = lhs.len()
        rhslength = rhs.len()
        lenlen = lhslength.len()
        if not _equal_int(lhslength, lenlen) and \
           not _equal_int(rhslength, lenlen) and \
           not _equal_int(lhslength, rhslength):
            raise ValueError(
                "Arrays are of incompatible lengths for operation.")

        # TODO: Perform broadcast automatically on the backend to improve performance
        if not _equal_int(lhslength, rhslength):
            broadcast_len = self.context().calc_broadcast_length([lhs, rhs])
            if not _equal_int(lhslength, broadcast_len):
                lhs = lhs.broadcast_value(broadcast_len)[0]
            if not _equal_int(rhslength, broadcast_len):
                rhs = rhs.broadcast_value(broadcast_len)[0]
        return self.context()._exec_command(f"{op} {handles_returned} {lhs.handle()} {rhs.handle()}")

    def _operator_single(self, op, param):
        param = self.context().promote(param)[0]
        return self.context()._exec_command(f"{op} 1 {param.handle()}")    

    def __init__(self, fc, **kwargs):
        super().__init__(fc=fc, **kwargs)

    def width(self):
        return 1

    def transmitter(self):
        return BuiltinTransmitter(self)

    def stub(self):
        return self.context().array(self.typecode())

    def flatten(self):
        return [self]

    def unflatten(self, data: list[ArrayIdentifier]):
        if type(data) is not list or len(data) != 1 or \
                type(data[0]) is not type(self) or \
                data[0].typecode() != self.typecode():
            # Note: this method does not attempt any value promotion.
            raise TypeError("Invalid input type for unflatten")
        self[:] = data[0]
        return self

    def copy(self):
        return self[:]

    def promote(self, typecode):
        if self.typecode() != typecode:
            raise TypeError("Cannot promote to the requested type.")
        return (self, False)

    def empty(self):
        self.set_length(0)

    def broadcast_value(self, length):
        if type(length) is int:
            length = self.context().array("i", 1, length)
        if type(length) is not IntArrayIdentifier:
            raise TypeError("Length parameter must be IntArrayIdentifier.")
        self.context().verify_context([length])
        lenlen = length.len()
        if _equal_int(self.len(), length):
            return (self, False)
        elif _equal_int(self.len(), lenlen):
            copy = self.copy()
            self.context()._exec_command(f"broadcast_value 0 {copy.handle()} {length.handle()}")
            return (copy, True)
        else:
            raise ValueError("Value not eligible for broadcasting.")

    def len(self):
        return self.context()._exec_command("len 1 " + self.handle())

    def __len__(self):
        with on(self.context().CoordinatorID):
            rc = self.context()._exec_command("pylen 0 " + self.handle())
        return rc

    def tolist(self): 
        with on(self.context().CoordinatorID):
            return self.context()._exec_command("tolist 0 " + self.handle())

    def set_length(self, new_length):
        new_length = self.context().promote(new_length, "i")
        self.context()._exec_command(f"setlength 0 {self.handle()} {new_length[0].handle()}")

    def __getitem__(self, key: Union(int, IntArrayIdentifier, slice)) -> ArrayIdentifier:
        # Gets items from array based on inputted key(s) - See Doco Section 5.2.2
        key = self._handle_key(key)
        return self.context()._exec_command(f"getitem 1 {self.handle()} {key.handle()}")

    def __delitem__(self, key: Union(int, IntArrayIdentifier, slice)):
        # Delete item(s) in-place from array based on inputted key(s) - See Doco Section 5.2.2
        k = self._handle_key(key)
        return self.context()._exec_command(f"delitem 0 {self.handle()} {k.handle()}")

    def lookup(self, key, default=None):
        key = self._handle_key(key)
        if default is None:
            return self.context()._exec_command(f"lookup 1 {self.handle()} {key.handle()}")
        else:
            default = self.context().promote(default, self.typecode())[0]
            return self.context()._exec_command(f"lookup 1 {self.handle()} {key.handle()} {default.handle()}")

    def __setitem__(self, key: Union(int, IntArrayIdentifier, slice), value: Union(int, float, ArrayIdentifier)):
        # Set item(s) in-place in array based on inputted key(s) and value(s) - See Doco Section 5.2.2
        if isinstance(key, slice) and key == slice(None, None, None):
            value = self.context().promote(value, self.typecode())[0]
            self.context()._exec_command(f"setitem 0 {self.handle()} {value.handle()}")
        else:
            key, value = self._handle_key_value(key, value)
            self.context()._exec_command(f"setitem 0 {self.handle()} {value.handle()} {key.handle()}")

    def __mux__(self, conditional, iffalse):
        conditional = self.context().promote(conditional, 'i')[0]
        iffalse = self.context().promote(iffalse, self.typecode())[0]
        broadcast_len = self.context().calc_broadcast_length([self, conditional, iffalse])

        conditional = conditional.broadcast_value(broadcast_len)[0]
        selfB = self.broadcast_value(broadcast_len)[0]
        iffalse = iffalse.broadcast_value(broadcast_len)[0]

        return self.context()._exec_command(f"mux 1 {conditional.handle()} {selfB.handle()} {iffalse.handle()}")

    def __rmux__(self, conditional, iftrue):
        conditional = self.context().promote(conditional, 'i')[0]
        iftrue = self.context().promote(iftrue, self.typecode())[0]
        broadcast_len = self.context().calc_broadcast_length([self, conditional, iftrue])

        conditional = conditional.broadcast_value(broadcast_len)[0]
        selfB = self.broadcast_value(broadcast_len)[0]
        iftrue = iftrue.broadcast_value(broadcast_len)[0]

        return self.context()._exec_command(f"mux 1 {conditional.handle()} {iftrue.handle()} {selfB.handle()}")

    def __eq__(self, other):
        if other is None:
            return False

        return self._operator("eq", self, other)

    def __ne__(self, other):
        if other is NotImplemented:
            return True

        return self._operator("ne", self, other)

    def _handle_key(self, key: Union(int, IntArrayIdentifier, slice)) -> IntArrayIdentifier:
        # Parses either Native Ints, IntArrayIdentifiers or Slices, and converts them to keys for arrays
        
        # Check if key is a slice object
        if type(key) is slice:
            start = key.start
            stop = key.stop
            if not isinstance(start, IntArrayIdentifier):
                if isinstance(start, int):
                    start = self.context().promote(start, "i")[0]
                elif start is None:
                    start = self.context().array('i', 1, 0)
                else:
                    raise ValueError("slice start should either be none or an int")
                
            if not isinstance(stop, IntArrayIdentifier):
                if isinstance(stop, int):
                    stop = self.context().promote(stop, "i")[0]
                elif stop is not None:
                    raise ValueError("slice stop should either be none or an int")
                
            command = f"slice_to_indices 1 {self.handle()} {start.handle()}"
            if stop is not None:
                command += f" {stop.handle()}"
            key = self.context()._exec_command(command)
        else:
            key = self.context().promote(key, "i")[0]
        return key

    def _handle_key_value(self, key, value):
        key = self._handle_key(key)
        value = self.context().promote(value, self.typecode())[0]
        key_len = key.len()
        value_len = value.len()

        if not _equal_int(value_len, key_len):
            broadcast_len = self.context().calc_broadcast_length([key, value])
            key = key.broadcast_value(broadcast_len)[0]
            value = value.broadcast_value(broadcast_len)[0]
        return key, value

    def serialise(self):
        return self.context()._exec_command(f"serialise 1 {self.handle()}")

    def deserialise(self, data):
        if not isinstance(data, BytearrayArrayIdentifier):
            raise ValueError(f"data must be of type BytearrayArrayIdentifier, not {type(data)}")
        else:
            self.context()._exec_command(f"deserialise 0 {self.handle()} {data.handle()}")

class FtilliteBuiltin(Identifier):
    def __init__(self, fc, handle):
        super().__init__(fc=fc)
        self._handle = handle
        weakref.finalize(self, fc.queue_for_deletion_append, handle, [i._num for i in self.scope()])

    def handle(self):
        return self._handle

    def __getstate__(self):
        state = {k: v for k, v in self.__dict__.items() if k != '_context'}
        state["_handle"] = self.handle()
        state["_scope"] = [x.num() for x in self._scope]
        state["registry"] = self.context()._save_handle(self.handle())
        state["_typecode"] = self.typecode()
        return state

    def __setstate__(self, state):
        # Python requires us to have a global "backend" object for this.
        self._context = backend.context()
        self.__dict__.update(state)
        if self._context == None:
            raise RuntimeError("Cannot load when not in a session.")
        self._scope = NodeSet([self.context().Nodes[x] for x in state["_scope"]])
        tc = typecode_list(state["_typecode"])
        if isinstance(self, ListMapIdentifier):
            tc = ['listmap'] + tc
        self._handle = self.context()._load_handle(state["_handle"],
                                                   state["registry"],
                                                   tc)


class IntArrayIdentifier(ArrayIdentifier, FtilliteBuiltin):
    def __init__(self, fc, handle):
        super(IntArrayIdentifier, self).__init__(fc=fc, handle=handle)

    def typecode(self):
        return 'i'

    def sametype(self, other):
        return type(other) is IntArrayIdentifier

    def promote(self, typecode):
        if self.typecode() == typecode:
            return (self, False)
        elif typecode in ['f', 'I']:
            return (self.astype(typecode), True)
        else:
            raise TypeError("Cannot promote to the requested type.")

    def astype(self, typecode):
        return self.context()._exec_command(f"astype 1 {self.handle()} {typecode}")

    def reduce_sum(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reducesum 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_isum(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reduceisum 0 {self.handle()} {value.handle()} {key.handle()}")

    def contains(self, other):
        return self.context()._exec_command(f"contains 1 {self.handle()} {other.handle()}")
        
    def cumsum(self):
        return self.context()._exec_command(f"cumsum 1 {self.handle()}")

    def index(self):
        return self.context()._exec_command(f"index 1 {self.handle()}")
        
    def reduce_max(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reducemax 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_imax(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reduceimax 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_min(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reducemin 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_imin(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reduceimin 0 {self.handle()} {value.handle()} {key.handle()}")

    def sorted(self):
        return self.context()._exec_command(f"sorted 1 {self.handle()}")

    def index_sorted(self, index=None):
        if index is not None:
            if not isinstance(index, IntArrayIdentifier):
                raise ValueError("index ")
            return self.context()._exec_command(f"indexsorted 1 {self.handle()} {index.handle()}")
        return self.context()._exec_command(f"indexsorted 1 {self.handle()}")

    def __iter__(self):
        with on(self.context().CoordinatorID):
            myvals = self.context()._exec_command("tolist 0 " + self.handle())
        for x in myvals:
            yield x

    def __lt__(self, other):
        return self._operator("lt", self, other)

    def __gt__(self, other):
        return self._operator("gt", self, other)

    def __le__(self, other):
        return self._operator("le", self, other)

    def __ge__(self, other):
        return self._operator("ge", self, other)

    def __abs__(self):
        return self._operator_single("abs", self)

    def __nonzero__(self):
        return self._operator_single("nonzero", self)

    def __pos__(self):
        raise NotImplementedError("Awaiting Golang implementation.")

    def __neg__(self):
        return self.context()._exec_command(f"neg 1 {self.handle()}")

    def __add__(self, other):
        return self._operator("add", self, other)

    def __iadd__(self, other):
        self[:] = self.__add__(other)
        return self

    def __radd__(self, other):
        return self._operator("add", other, self)

    def __sub__(self, other):
        return self._operator("sub", self, other)

    def __isub__(self, other):
        self[:] = self.__sub__(other)
        return self

    def __rsub__(self, other):
        return self._operator("sub", other, self)
        
    def __mul__(self, other):
        # Handle specific case for i * E
        lhs = self
        if isinstance(other, Ed25519ArrayIdentifier):
            lhs = lhs.astype('I')
        return self._operator("mul", lhs, other, type_checker=lambda lhs, rhs: \
            type(lhs) in [Ed25519IntArrayIdentifier])

    def __imul__(self, other):
        self[:] = self.__mul__(other)
        return self

    def __rmul__(self, other):
        return self._operator("mul", self, other)

    def __floordiv__(self, other):
        return self._operator("floordiv", self, other)

    def __ifloordiv__(self, other):
        self[:] = self.__floordiv__(other)
        return self

    def __rfloordiv__(self, other):
        return self._operator("floordiv", other, self)

    def __truediv__(self, other):
        # Note: integer division is handled the same way as floor division
        return self._operator("floordiv", self, other)

    def __itruediv__(self, other):
        self[:] = self.__floordiv__(other)
        return self

    def __rtruediv__(self, other):
        return self._operator("truediv", other, self)

    def __mod__(self, other):
        return self._operator("mod", self, other)

    def __imod__(self, other):
        self[:] = self.__mod__(other)
        return self

    def __rmod__(self, other):
        return self._operator("mod", other, self)

    def __divmod__(self, other):
        return self._operator("divmod", self, other, 2)

    def __rdivmod__(self, other):
        return self._operator("divmod", self, other, 2)

    def __pow__(self, other):
        return self._operator("pow", self, other)

    def __ipow__(self, other):
        self[:] = self.__pow__(other)
        return self

    def __rpow__(self, other):
        return self._operator("pow", other, self)

    def __lshift__(self, other):
        return self._operator("lshift", self, other)

    def __ilshift__(self, other):
        self[:] = self.__lshift__(other)
        return self

    def __rlshift__(self, other):
        return self._operator("lshift", other, self)

    def __rshift__(self, other):
        return self._operator("rshift", self, other)

    def __irshift__(self, other):
        self[:] = self.__rshift__(other)
        return self

    def __rrshift__(self, other):
        return self._operator("rshift", other, self)

    def __and__(self, other):
        return self._operator("and", self, other)

    def __iand__(self, other):
        self[:] = self.__and__(other)
        return self

    def __rand__(self, other):
        return self._operator("and", other, self)

    def __or__(self, other):
        return self._operator("or", self, other)

    def __ior__(self, other):
        self[:] = self.__or__(other)
        return self

    def __ror__(self, other):
        return self._operator("or", other, self)

    def __xor__(self, other):
        return self._operator("xor", self, other)

    def __ixor__(self, other):
        self[:] = self.__xor__(other)
        return self

    def __rxor__(self, other):
        return self._operator("xor", other, self)

    def __invert__(self):
        return self._operator_single("invert", self)

    def __nearest__(self):
        return self._operator_single("nearest", self)


class FloatArrayIdentifier(ArrayIdentifier, FtilliteBuiltin):
    def __init__(self, fc, handle):
        super(FloatArrayIdentifier, self).__init__(fc=fc, handle=handle)

    def typecode(self):
        return 'f'

    def sametype(self, other):
        return type(other) is FloatArrayIdentifier

    def astype(self, typecode):
        return self.context()._exec_command(f"astype 1 {self.handle()} {typecode}")

    def reduce_sum(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reducesum 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_isum(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reduceisum 0 {self.handle()} {value.handle()} {key.handle()}")

    def contains(self, other):
        return self.context()._exec_command(f"contains 1 {self.handle()} {other.handle()}")

    def cumsum(self):
        return self.context()._exec_command(f"cumsum 1 {self.handle()}")

    def index(self):
        return self.context()._exec_command(f"index 1 {self.handle()}")
    
    def reduce_max(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reducemax 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_imax(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reduceimax 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_min(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reducemin 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_imin(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reduceimin 0 {self.handle()} {value.handle()} {key.handle()}")

    def sorted(self):
        self.context()._exec_command(f"sorted 1 {self.handle()}")

    def index_sorted(self, index=None):
        if index is not None:
            if not isinstance(index, IntArrayIdentifier):
                raise ValueError("index ")
            return self.context()._exec_command(f"indexsorted 1 {self.handle()} {index.handle()}")
        return self.context()._exec_command(f"indexsorted 1 {self.handle()}")

    def __iter__(self):
        with on(self.context().CoordinatorID):
            myvals = self.context()._exec_command("tolist 0 " + self.handle())
        for x in myvals:
            yield x

    def __lt__(self, other):
        return self._operator("lt", self, other)

    def __gt__(self, other):
        return self._operator("gt", self, other)

    def __le__(self, other):
        return self._operator("le", self, other)

    def __ge__(self, other):
        return self._operator("ge", self, other)

    def __abs__(self):
        return self._operator_single("abs", self)

    def __pos__(self):
        raise NotImplementedError("Awaiting Golang implementation.")

    def __neg__(self):
        return self.context()._exec_command(f"neg 1 {self.handle()}")

    def __floor__(self):
        return self._operator_single("floor", self)

    def __ceil__(self):
        return self._operator_single("ceil", self)

    def __round__(self):
        return self._operator_single("round", self)

    def __add__(self, other):
        return self._operator("add", self, other)

    def __iadd__(self, other):
        self[:] = self.__add__(other)
        return self

    def __radd__(self, other):
        return self._operator("add", self, other)

    def __sub__(self, other):
        return self._operator("sub", self, other)

    def __isub__(self, other):
        self[:] = self.__sub__(other)
        return self

    def __rsub__(self, other):
        return self._operator("sub", other, self)
        
    def __mul__(self, other):
        return self._operator("mul", self, other)

    def __imul__(self, other):
        self[:] = self.__mul__(other)
        return self

    def __rmul__(self, other):
        return self._operator("mul", self, other)

    def __truediv__(self, other):
        return self._operator("truediv", self, other)

    def __itruediv__(self, other):
        self[:] = self.__truediv__(other)
        return self

    def __rtruediv__(self, other):
        return self._operator("truediv", other, self)
    
    def __pow__(self, other: Union[int, IntArrayIdentifier]):
        # rhs must be of type int
        return self._operator("pow", self, other, type_checker=lambda lhs, rhs: \
            type(rhs) in [int, IntArrayIdentifier])

    def __ipow__(self, other):
        self[:] = self.__pow__(other)
        return self

    def __exp__(self):
        return self._operator_single("exp", self)

    def __log__(self):
        return self._operator_single("log", self)

    def __sin__(self):
        return self._operator_single("sin", self)

    def __cos__(self):
        return self._operator_single("cos", self)


class Ed25519IntArrayIdentifier(ArrayIdentifier, FtilliteBuiltin):
    def __init__(self, fc, handle):
        super(Ed25519IntArrayIdentifier, self).__init__(fc=fc, handle=handle)

    def typecode(self):
        return 'I'

    def sametype(self, other):
        return type(other) is Ed25519IntArrayIdentifier

    def astype(self, typecode):
        return self.context()._exec_command(f"astype 1 {self.handle()} {typecode}")

    def len(self):
        return self.context()._exec_command("len 1 " + self.handle())

    def reduce_sum(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reducesum 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_isum(self, key, value):
        raise NotImplementedError("Awaiting Golang implementation.")

    def contains(self, other):
        return self.context()._exec_command(f"contains 1 {self.handle()} {other.handle()}")

    def cumsum(self):
        return self.context()._exec_command(f"cumsum 1 {self.handle()}")

    def index(self):
        return self.context()._exec_command(f"index 1 {self.handle()}")

    def __pos__(self):
        raise NotImplementedError("Awaiting Golang implementation.")

    def __neg__(self):
        return self.context()._exec_command(f"neg 1 {self.handle()}")

    def __add__(self, other):
        return self._operator("add", self, other)

    def __iadd__(self, other):
        self[:] = self.__add__(other)
        return self

    def __radd__(self, other):
        return self._operator("add", self, other)

    def __sub__(self, other):
        return self._operator("sub", self, other)

    def __isub__(self, other):
        self[:] = self.__sub__(other)
        return self

    def __rsub__(self, other):
        return self._operator("sub", other, self)
        
    def __mul__(self, other):
        return self._operator("mul", self, other, type_checker=lambda lhs, rhs: \
            type(rhs) in [Ed25519ArrayIdentifier, Ed25519IntArrayIdentifier])

    def __imul__(self, other):
        self[:] = self.__mul__(other)
        return self

    def __rmul__(self, other):
        return self._operator("mul", self, other)

    def __floordiv__(self, other):
        return self._operator("floordiv", self, other)

    def __ifloordiv__(self, other):
        self[:] = self.__floordiv__(other)
        return self

    def __rfloordiv__(self, other):
        return self._operator("floordiv", other, self)

    def __truediv__(self, other):
        return self._operator("truediv", self, other)

    def __itruediv__(self, other):
        self[:] = self.__truediv__(other)
        return self

    def __rtruediv__(self, other):
        return self._operator("truediv", other, self)
    
    def __pow__(self, other: Union[int, IntArrayIdentifier]):
        # rhs must be of type int
        return self._operator("pow", self, other, type_checker=lambda lhs, rhs: \
            type(rhs) in [int, IntArrayIdentifier])

    def __ipow__(self, other):
        self[:] = self.__pow__(other)
        return self
class Ed25519ArrayIdentifier(ArrayIdentifier, FtilliteBuiltin):
    def __init__(self, fc, handle):
        super(Ed25519ArrayIdentifier, self).__init__(fc=fc, handle=handle)

    def typecode(self):
        return 'E'

    def sametype(self, other):
        return type(other) is Ed25519ArrayIdentifier

    def astype(self, typecode):
        return self.context()._exec_command(f"astype 1 {self.handle()} {typecode}")

    def len(self):
        return self.context()._exec_command("len 1 " + self.handle())

    def reduce_sum(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reducesum 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_isum(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reduceisum 0 {self.handle()} {value.handle()} {key.handle()}")

    def contains(self, other):
        return self.context()._exec_command(f"contains 1 {self.handle()} {other.handle()}")

    def cumsum(self):
        return self.context()._exec_command(f"cumsum 1 {self.handle()}")

    def index(self):
        return self.context()._exec_command(f"index 1 {self.handle()}")

    def ed_folded(self):
        return self.context()._exec_command("ed_folded 1 " + self.handle())

    def ed_affine(self):
        return self.context()._exec_command("ed_affine 1 " + self.handle())
    
    def __pos__(self):
        raise NotImplementedError("Awaiting Golang implementation.")

    def __neg__(self):
        return self.context()._exec_command(f"neg 1 {self.handle()}")

    def __add__(self, other):
        return self._operator("add", self, other)

    def __iadd__(self, other):
        self[:] = self.__add__(other)
        return self

    def __sub__(self, other):
        return self._operator("sub", self, other)

    def __isub__(self, other):
        self[:] = self.__sub__(other)
        return self
        
    def __mul__(self, other):
        # Handle specific case for E * i 
        other = self._handle_native_type(other)
        if isinstance(other, IntArrayIdentifier):
            other = other.astype('I')
        return self._operator("mul", self, other, type_checker=lambda lhs, rhs: \
            type(rhs) in [Ed25519IntArrayIdentifier])

    def __imul__(self, other):
        self[:] = self.__mul__(other)
        return self

    def __rmul__(self, other):
        return self.__mul__(other)


class BytearrayArrayIdentifier(ArrayIdentifier, FtilliteBuiltin):
    def __init__(self, fc, handle, typecode):
        super(BytearrayArrayIdentifier, self).__init__(fc=fc, handle=handle)
        self._typecode = typecode
        self._size = int(typecode[1:])

    def typecode(self):
        return self._typecode

    def sametype(self, other):
        return type(other) is BytearrayArrayIdentifier and \
            self.typecode() == other.typecode()

    def astype(self, typecode):
        return self.context()._exec_command(f"astype 1 {self.handle()} {typecode}")

    def reduce_sum(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reducesum 0 {self.handle()} {value.handle()} {key.handle()}")

    def reduce_isum(self, key, value):
        key, value = self._handle_key_value(key, value)
        self.context()._exec_command(f"reduceisum 0 {self.handle()} {value.handle()} {key.handle()}")

    def contains(self, other):
        return self.context()._exec_command(f"contains 1 {self.handle()} {other.handle()}")

    def cumsum(self):
        return self.context()._exec_command(f"cumsum 1 {self.handle()}")

    def index(self):
        return self.context()._exec_command(f"index 1 {self.handle()}")

    def serialise(self):
        return self

    def deserialise(self, data):
        if not isinstance(data, BytearrayArrayIdentifier):
            raise ValueError(f"data must be of type BytearrayArrayIdentifier, not {type(data)}")
        else:
            self[:] = data.copy()

    def concat(self, other):
        return self._operator("concat", self, other)

    def byteproject(self, size, mapping):
        mKeys = []
        mValues = []
        for k in mapping:
            mKeys.append(k)
            mValues.append(mapping[k])

        keys = self.context().array("i", mKeys)
        values = self.context().array("i", mValues)

        return self.context()._exec_command(f"byteproject 1 {self.handle()} {size} {keys.handle()} {values.handle()}")

    def ed_folded_project(self):
        return self.context()._exec_command("ed_folded_project 1 " + self.handle())

    def ed_affine_project(self):
        return self.context()._exec_command("ed_affine_project 1 " + self.handle())

    def size(self):
        return self._size

    def __lshift__(self, other):
        return self._operator("lshift", self, other)

    def __ilshift__(self, other):
        self[:] = self.__lshift__(other)
        return self

    def __rlshift__(self, other):
        return self._operator("lshift", other, self)

    def __rshift__(self, other):
        return self._operator("rshift", self, other)

    def __irshift__(self, other):
        self[:] = self.__rshift__(other)
        return self

    def __rrshift__(self, other):
        return self._operator("rshift", other, self)

    def __and__(self, other):
        return self._operator("and", self, other)

    def __iand__(self, other):
        self[:] = self.__and__(other)
        return self

    def __rand__(self, other):
        return self._operator("and", other, self)

    def __or__(self, other):
        return self._operator("or", self, other)

    def __ior__(self, other):
        self[:] = self.__or__(other)
        return self

    def __ror__(self, other):
        return self._operator("or", other, self)

    def __xor__(self, other):
        return self._operator("xor", self, other)

    def __ixor__(self, other):
        self[:] = self.__xor__(other)
        return self

    def __rxor__(self, other):
        return self._operator("xor", other, self)

    def __invert__(self):
        return self._operator_single("invert", self)

class ListMapIdentifier(FtilliteBuiltin):
    def __init__(self, fc, typecode, handle):
        super(ListMapIdentifier, self).__init__(fc=fc, handle=handle)
        # Note: Empty typecode ListMaps are not allowed. Width must be >= 1.
        self._handle = handle
        self._typecode = typecode

    def typecode(self) -> str:
        # Returns a string representation of the listmap key typecodes - See Doco Section 5.3.3
        return self._typecode

    def transmitter(self) -> BuiltinTransmitter:
        # Returns transmitter object - See Doco Section 5.3.3
        return BuiltinTransmitter(self)

    def stub(self) -> ListMapIdentifier:
        # Returns empty listmap with same typecodes as self - See Doco Section 5.3.3
        return self.context().listmap(self.typecode())

    def sametype(self, other: ListMapIdentifier) -> bool:
        # Checks for type equality between two listmaps - See Doco Section 5.3.3
        return type(other) is ListMapIdentifier and \
            self.typecode() == other.typecode()

    def flatten(self) -> list(ArrayIdentifier):
        # Flattens listmap into array of key composites - See Doco Section 5.3.3
        return self.keys()

    def unflatten(self, data: list(ArrayIdentifier)) -> ListMapIdentifier:
        # Replace listmap keys with data keys - See Doco Section 5.3.3
        if type(data) is not list:
            raise TypeError("Invalid input type for unflatten")
        self[:] = self.context().listmap(data, order="pos")
        return self

    def copy(self) -> ListMapIdentifier:
        # Returns copy of listmap - See Doco Section 5.3.3
        return self.context()._exec_command(f"listmap_copy 1 {self.handle()}")

    def len(self) -> IntArrayIdentifier:
        # Returns singleton array identifier of length - See Doco Section 5.3.3
        return self.keys()[0].len()

    def listmap_contains_all_keys(self, keys: list(ArrayIdentifier)) -> bool:
        # Util method to check that all keys provided exist in the listmap
        if not isinstance(keys, list):
            raise ValueError("keys is not of type list")
        for k in keys:
            if not isinstance(k, ArrayIdentifier):
                raise ValueError("keys must be a list of array identifiers")

        # TODO: find a better way using sum for the below check
        return _equal_int(self.contains(keys), self.context().array('i', keys[0].len(), 1))

    def listmap_contains_no_keys(self, keys: list(ArrayIdentifier)) -> bool:
        # Util method to check that no keys provided exist in the listmap
        if not isinstance(keys, list):
            raise ValueError("keys is not of type list")
        for k in keys:
            if not isinstance(k, ArrayIdentifier):
                raise ValueError("keys must be a list of array identifiers")

        return _equal_int(self.contains(keys), self.context().array('i', keys[0].len(), 0))

    def __setitem__(self, key, data: ListMapIdentifier):
        # Replaces values of a listmap in-place - See Doco Section 5.3.3
        
        if key != slice(None, None, None):
            raise KeyError("Only the slice ':' is supported.")
        # The "isinstance(data, ListMapIdentifier)" can be switched to
        # "type(data) is not ListMapIdentifier" if we want to exclude
        # inherited types.
        if not isinstance(data, ListMapIdentifier):
            raise TypeError(f"Argument type incompatible for assignment. date type != ListMapIdentifier, instead is {type(data)}")
        elif not self.sametype(data):
            raise TypeError(f"Argument type incompatible for assignment. Listmap type != data type, {self.typecode()} != {data.typecode()}")
        data = data.copy()
        self.context()._exec_command(f"listmap_setitems 0 {self.handle()} {data.handle()}")

    def todict(self) -> dict[list(Union[int,float]), int]:
        # Converts listmap on Coordinator node to native python dictionary - See Doco Section 5.3.3
        tcs = typecode_list(self.typecode())
        if any([tc not in ['i', 'f'] for tc in tcs]):
            raise TypeError("operation can only be performed for listmaps composed of 'i' and 'f' key composites")
        keys_arrays = [k.tolist() for k in self.keys()]
        dict_keys = []
        for i in range(len(keys_arrays[0])):
            tmp = ()
            for j in range(len(keys_arrays)):
                tmp = tmp + (keys_arrays[j][i],)
            dict_keys.append(tmp)
        dict_values = self[keys_arrays].tolist()
        return { dict_keys[i]: dict_values[i] for i in range(len(dict_keys)) }

    def __len__(self):
        # Returns number of keys in listmap on Coordinator node - See Doco Section 5.3.3
        with on(self.context().CoordinatorID):
            return len(self.keys()[0]) if self.width() > 0 else 0

    def keys(self) -> list(ArrayIdentifier):
        # Returns the keys of the listmap
        keys = self.context()._exec_command(f"listmap_keys {self.width()} {self.handle()}")
        if not isinstance(keys, list):
            return [keys]
        else:
            return keys
        
    def handle_item_input(self, item, listmap_allowed=True):
        # Util method to handle key and listmap inputs for listmap methods
        if not isinstance(item, list) and not isinstance(item, ListMapIdentifier):
            raise ValueError(f"item is not of type list or ListMapIdentifier (is {type(item)})")
        elif not listmap_allowed and isinstance(item, ListMapIdentifier): 
            raise ValueError(f"item cannot be of type ListMapIdentifier (is {type(item)})")
        elif isinstance(item, ListMapIdentifier):
            item = item.keys()

        if len(item) != self.width():
            raise ValueError(f"item width does not match listmap key width, {len(item)} != {self.width()}")

        broadcast_len = self.context().calc_broadcast_length(item)
        for i, k in enumerate(item):
            item[i] = k.broadcast_value(broadcast_len)[0]
        return item

    def _get(self, keys: list(ArrayIdentifier), keys_exist: bool = True, default: int = -1) -> IntArrayIdentifier:
        # Returns the values for keys in the listmap - See Doco Section 5.3.3
        keys = self.handle_item_input(keys, listmap_allowed=False)
        if keys_exist and not self.listmap_contains_all_keys(keys):
            raise ValueError("one or many keys are not in listmap")
        if not keys_exist:
            default = self.context().promote(default, 'i')[0]
            default_val = default.handle()
        else:
            default_val = 0
        return self.context()._exec_command(f"listmap_getitem 1 {self.handle()} {default_val} {' '.join([k.handle() for k in keys])}")

    def __getitem__(self, keys: list(ArrayIdentifier)) -> IntArrayIdentifier:      
        return self._get(keys)
    
    def lookup(self, keys: list(ArrayIdentifier), default: int = -1) -> IntArrayIdentifier:
        return self._get(keys, False, default)
        
    def contains(self, items: list(ArrayIdentifier)) -> IntArrayIdentifier:
        items = self.handle_item_input(items, listmap_allowed=False)
        return self.context()._exec_command(f"listmap_contains 1 {self.handle()} {' '.join([k.handle() for k in items])}")

    def _add(self, items: Union[list(ArrayIdentifier), ListMapIdentifier], allow_duplicate_items: bool=False, allow_existing_keys: bool = False) -> Tuple(list, IntArrayIdentifier):
        # Adds items in-place to listmap - See Doco Section 5.3.3
        items = self.handle_item_input(items)
        if not allow_duplicate_items and not self.context().keys_unique(items):
            raise ValueError("duplicate keys found in items")
        if not allow_existing_keys and not self.listmap_contains_no_keys(items):
            raise ValueError("keys in items found in listmap")

        ignore_errors = int(allow_duplicate_items or allow_existing_keys)
        res = self.context()._exec_command(f"listmap_additem {len(items) + 1} {self.handle()} {ignore_errors} {'_'.join([k.handle() for k in items])}")
        if isinstance(res, list) and len(res) == self.width() + 1:
            keys = res[0:-1]
            value = res[-1]
            return keys, value
        else:
            raise ValueError("incorrect result returned from server")

    def add_items(self, items: Union[list(ArrayIdentifier), ListMapIdentifier]) -> Tuple(list, IntArrayIdentifier):
        return self._add(items)

    def merge_items(self, items: Union[list(ArrayIdentifier), ListMapIdentifier]) -> Tuple(list, IntArrayIdentifier):
        return self._add(items, True, True)

    def _remove(self, items: Union[list(ArrayIdentifier), ListMapIdentifier], allow_duplicate_items: bool = False, allow_keys_not_present: bool = False)  -> Tuple(list, IntArrayIdentifier, IntArrayIdentifier):
        # Remove items in-place to listmap - See Doco Section 5.3.3
        items = self.handle_item_input(items)
        if not allow_duplicate_items and not self.context().keys_unique(items):
            raise ValueError("duplicate keys found in items")
        if not allow_keys_not_present and not self.listmap_contains_all_keys(items):
            raise ValueError("one or more keys in items not found in listmap")
        
        ignore_errors = int(allow_duplicate_items or allow_keys_not_present)
        
        res = self.context()._exec_command(f"listmap_removeitem {len(items) + 2} {self.handle()} {ignore_errors} {'_'.join([k.handle() for k in items])}")
        if isinstance(res, list) and len(res) == self.width() + 2:
            moved_keys = res[:-2]
            old_values = res[-2]
            new_values = res[-1]
            return (moved_keys, old_values, new_values)
        else:
            raise ValueError("incorrect result returned from server")    

    def remove_items(self, items: Union[list(ArrayIdentifier), ListMapIdentifier]) -> Tuple(list, IntArrayIdentifier, IntArrayIdentifier):
        return self._remove(items)
        
    def discard_items(self, items: Union[list(ArrayIdentifier), ListMapIdentifier]) -> Tuple(list, IntArrayIdentifier, IntArrayIdentifier):
        return self._remove(items, True, True)
    
    def intersect_items(self, items: Union[list, ListMapIdentifier]) -> ListMapIdentifier:
        # Provides a new listmap with the intersection between keys of two listmaps - See Doco Section 5.3.3
        items = self.handle_item_input(items)
        return self.context()._exec_command(f"listmap_intersectitem 1 {self.handle()} {'_'.join([k.handle() for k in items])}")


class Transmitter(ManagedObject):
    def __init__(self, fc):
        super().__init__(fc)


class BuiltinTransmitter(Transmitter):
    def __init__(self, obj):
        super().__init__(obj.context())
        self._type = type(obj)
        self._typecode = obj.typecode()
        if not isinstance(obj, FtilliteBuiltin):
            raise TypeError("Incorrect transmitter type for object.")
        if isinstance(obj, ArrayIdentifier):
            self._opcode = "array"
        elif type(obj) is ListMapIdentifier:
            self._opcode = "listmap"
        else:
            raise TypeError("Incorrect transmitter type for object.")

    def transmit(self, inmap):
        keyscope = set(inmap.keys())
        if not keyscope.issubset(self.context().scope()):
            raise ValueError(
                "Cannot perform transmit in this execution scope.")
        # Transmitter type match must be exact.
        for k, v in inmap.items():
            if type(k) is not Node:
                raise TypeError("Incorrect key type for transmitter.")
            if type(v) is not self._type or not isinstance(v, FtilliteBuiltin):
                raise TypeError("Incorrect value type for transmitter.")
            if not v.scope().issubset(self.context().scope()):
                raise ValueError("Execution scope too narrow for operation.")
        return self.context().transmit(self._opcode, self._typecode,
                                       " ".join([str(k.num()) + " " + v.handle() for k, v in inmap.items()]))


def _equal_int(val1, val2):
    try:
        rc = val1.context()._exec_command("equalint 0 " + val1.handle() + " " +
                                          val2.handle())
    except TypeError:
        return "bool 0"
    return rc


def verify(condition: IntArrayIdentifier, raise_exception = True) -> bool:
    if not isinstance(condition, IntArrayIdentifier):
        raise TypeError("condition parameter not of type IntArrayIdentifier")

    result = False
    
    try:
        result = condition.context()._exec_command("verify 0 " + condition.handle())
    except TypeError:
        result = False

    if not result and raise_exception:
       raise AssertionError("Verify failed: handle " + condition.handle())

    return result

def transmit(dictionary):
    if len(dictionary) == 0:
        raise ValueError("Cannot transmit an empty dictionary.")
    for e in dictionary.values():
        break
    t = e.transmitter()
    return t.transmit(dictionary)


def sha3_256(data):
    return data.context()._exec_command(f'sha3_256 1 {data.handle()}')


def ecdsa256_sign(data, priv_key):
    return data.context()._exec_command(f'ecdsa256_sign 1 {data.handle()} {priv_key.handle()}')


def ecdsa256_verify(data, signature, pub_key):
    return data.context()._exec_command(f'ecdsa256_verify 1 {data.handle()} {signature.handle()} {pub_key.handle()}')


def aes256_encrypt(data, key):
    return data.context()._exec_command(f'aes256_encrypt 1 {data.handle()} {key.handle()}')


def aes256_decrypt(data, key):
    return data.context()._exec_command(f'aes256_decrypt 1 {data.handle()} {key.handle()}')


def grain128aeadv2(key, iv, size, length):
    v_size = key.context().promote(size, "i")

    return key.context()._exec_command(f'grain128aeadv2 1 {key.handle()} {iv.handle()} {v_size[0].handle()} {length.handle()}')


def rsa3072_encrypt(data, pub_key):
    return data.context()._exec_command(f'rsa3072_encrypt 1 {data.handle()} {pub_key.handle()}')


def rsa3072_decrypt(data, priv_key):
    return data.context()._exec_command(f'rsa3072_decrypt 1 {data.handle()} {priv_key.handle()}')


def mux(conditional, iftrue, iffalse):
    if hasattr(iftrue, "__mux__"):
        return iftrue.__mux__(conditional, iffalse)
    elif hasattr(iffalse, "__rmux__"):
        return iffalse.__rmux__(conditional, iftrue)
    elif set([type(iftrue), type(iffalse)]).issubset(set([int, float])):
        return (conditional != 0) * iftrue + (conditional == 0) * iffalse
    else:
        raise TypeError("Operands do not support mux function.")


def get_context(params):
    e = None
    for e in params:
        if isinstance(e, ManagedObject):
            break
    if not isinstance(e, ManagedObject):
        raise TypeError("No ManagedObjects found in argument.")
    e.context().verify_context(params)
    return e.context()


def ceil(x):
    return x.__ceil__()


def floor(x):
    return x.__floor__()


def round(x):
    return x.__round__()


def nearest(x):
    return x.__nearest__()


def exp(x):
    return x.__exp__()


def log(x):
    return x.__log__()


def sin(x):
    return x.__sin__()


def cos(x):
    return x.__cos__()

def typecode_list(tc):
    return re.split(r'([ifIE]|b[1-9][0-9]*)\s*', tc)[1::2]

# Support for "massop()" given below is a do-nothing placeholder.

class massop:
    def __enter__(self):
        pass

    def __exit__(self, exc_type, exc_value, traceback):
        pass
