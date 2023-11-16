# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################


# Demo for the ftillite Python backend
# ====================================
#
# Revision history:
#
##################################################################

# This is a mock implementation of an ftillite compute manager.
# It implements "ComputeManager.run_command", which accepts a command
# string from the FTILContext object, triggers back-end operations at the
# segment managers, collates responses from the segment managers, handles
# any errors they report, and (if there are no errors) returns a string
# result.
#
# Given that we plan the compute manager to be implemented as a Python
# in-process solution also in the final deliverable, it is not strictly
# necessary for communication between the FTILContext and the compute manager
# to be done via strings. Our solution uses some string communication where
# convenient, and direct function calls when not, but the communication and
# the calls are structured in a way that if in future we want to separate
# the compute manager to a separate process, this will be trivial to do.
#
# This mock implementation does not simulate the initial setting up of the
# network. This is done here using direct function calls that create and set
# up native Python objects.


import random
from ftillite.segment_client import SegmentClient
from ftillite.segment_node import SegmentNode
import threading
import logging
from typing import Callable, List
class ComputeManager:

    def add_sm(self, name):
        self.segment_clients.append(SegmentClient(
            self, 
            name, 
            len(self.segment_clients), 
            self.rabbitmq_conf,
            self.segment_client_logger))
        self.sm_name.append(name)

    def context(self):
        return self._context

    def __init__(self, rabbitmq_conf=None, logger=None, segment_client_logger=None, peer_names=[]):
        self.variable_registry = {}
        self._context = None
        self.segment_clients = []
        self.sm_name = []
        self.rabbitmq_conf = rabbitmq_conf
        self.logger = logger
        self.segment_client_logger = segment_client_logger
        self.del_await_resp = rabbitmq_conf.get('del_await_resp', True)
        self.peer_names = peer_names

        # The first to be added will be the coordinator node.
        for p in peer_names:
            self.add_sm(p)

    def _print(self, value, loglevel=logging.INFO):
        if self.logger is not None:
            self.logger.log(loglevel, f"COMPUTE MANAGER - {value}")

    def _get_new_handle(self):
        cond = True
        while cond:
            r = str(random.getrandbits(63))
            cond = r in self.variable_registry
        # Creating a dummy object to avoid race conditions.
        self.variable_registry[r] = None
        return r

    def _extract_singleton(self, node, address):
        self._print(f"SINGLETON: self.variable_store[{node}][{address}]")
        valuearray = self.variable_store[node][address]
        if len(valuearray) != 1:
            raise ValueError("Value is not a singleton.")
        return valuearray[0]

    def init(self, context):
        # This is a direct function call that, if the compute manager is
        # separated to another process, will have to be replaced by text
        # communication. There will still need to be a local Python object
        # maintaining the reference to "context", though, because that is needed
        # for implementing "FTILContext.load".
        if self.context() != None:
            raise RuntimeError("Cannot hold concurrent sessions.")
        node = []
        for i in range(len(self.segment_clients)):
            # This will be MyID
            self.segment_clients[i].run_command(f"newilist 0 {i}")
            node.append(SegmentNode(self.segment_clients[i].run_command(f"init 0 {i}")))
        for i in range(len(self.segment_clients)):
            self.segment_clients[i].run_command(f"netinit " + " ".join(str(x) for x in node))
        self.variable_registry["0"] = ("array", "i",
                                       set(range(len(self.segment_clients))))
        # Values are: (type, typecode, scope)
        self._context = context

        self.gpuenabled = all(seg.gpu == "gpu" for seg in node)
        # The first of the returned values is the coordinator.
        return [(i, self.sm_name[i]) for i in range(len(self.sm_name))]

    def clear_variable_registry(self):
        tmp = self.variable_registry.copy()
        
        for v in tmp:
            self.del_handle(v)

    def del_handle(self, handle: str):
        # Handles single variable deletion
        (_, _, varscope) = self.variable_registry[handle]
        def tmp(i): 
            self.segment_clients[i].run_command(f"del 0 {handle}", response_required=self.del_await_resp)

        # Run concurrent deletion
        self.process_concurrently(tmp, varscope)
            
        del self.variable_registry[handle]

    def del_handles(self, handle_dict: dict):
        # Handles bulk variable deletion
        def tmp(i): 
            self.segment_clients[i].run_command(f"del 0 {' '.join(handle_dict[i])}", response_required=self.del_await_resp)

        # Run concurrent deletion
        self.process_concurrently(tmp, handle_dict)

        handles = list(set([y for x in handle_dict.values() for y in x]))
        for h in handles:
            del self.variable_registry[h]
        
    def _code(self, handle: str):
        return self.variable_registry[handle][0] + " " \
            + self.variable_registry[handle][1] + " " + handle

    def transmit(self, opcode, dtype, request):
        # To use textual interface if separating processes.
        request = request.split(" ")
        rc = {}
        transmit_pairs = []
        transmit_commands = {}
        while len(request) != 0:
            if self.variable_registry[request[1]][:2] != (opcode, dtype):
                raise TypeError("Incorrect type for transmitting.")
            node = int(request[0])

            for i in self.variable_registry[request[1]][2]:
                if i not in rc:
                    rc[i] = self._get_new_handle()
                    self.variable_registry[rc[i]] = (opcode, dtype, set())
                pair = (i, node)
                transmit_pairs.append(pair)
                transmit_commands[pair] = (i, f"transmit {node} {rc[i]} {request[1]} {dtype} {opcode}")
                self.variable_registry[rc[i]][2].add(node)

            del request[:2]

        def transmit_single(pair):
            node_id, node_cmd = transmit_commands[pair]
            retval = self.segment_clients[node_id].run_command(node_cmd)
            if retval != "ack":
                raise RuntimeError(
                    "Unexpected error in transmit: " + retval)

        for i, phase in enumerate(self.transmit_order(transmit_pairs)):
            self.process_concurrently(transmit_single, phase)

        return "map " + " ".join([str(k) + " " + self._code(v)
                                  for k, v in rc.items()])

    def transmit_order(self, combs):
        combs_new = sorted(combs, key=lambda tup: (tup[0], tup[1]))
        transmit_order = []
        processed = []
        senders = {}
        for i in combs_new:
            if i[0] not in senders:
                senders[i[0]] = []
            senders[i[0]].append(i[1])
        tmp_transmit_order = [1]
        while len(tmp_transmit_order) != 0:
            tmp_transmit_order = []
            receiving_tmp = []
            for i, v in senders.items():
                for j in range(len(v)):
                    idx = (i + j + 1) % len(v)
                    if (i, v[idx]) not in processed and v[idx] not in receiving_tmp:
                        processed.append((i, v[idx]))
                        tmp_transmit_order.append((i, v[idx]))
                        receiving_tmp.append(v[idx])
                        break
            if len(tmp_transmit_order) > 0:
                transmit_order.append(tmp_transmit_order)   
        return transmit_order

    def start_remote_save(self, destination):
        # This method is not reentrant. There is a race condition here.
        if hasattr(self, "save_destination"):
            raise RuntimeError("Concurrency is not allowed.")
        self.save_destination = destination
        rcmap = {}
        for s in self.segment_clients:
            rcmap[s] = s.run_command(f"startsave {self.save_destination}")
        self._check_errors(rcmap, [])

    def save_handle(self, handle):
        if not hasattr(self, "save_destination"):
            raise RuntimeError("Can only call this when save is in progress.")
        (vartype, _, _) = self.variable_registry[handle]
        rcmap = {}
        for s in self.variable_registry[handle][2]:
            rcmap[s] = self.segment_clients[s].run_command(f"save {handle} {vartype}")
        self._check_errors(rcmap, [])
        return self.variable_registry[handle]

    def finish_remote_save(self):
        if not hasattr(self, "save_destination"):
            raise RuntimeError("Can only call this when save is in progress.")
        rcmap = {}
        for s in self.segment_clients:
            rcmap[s] = s.run_command(f"finishsave")
        self._check_errors(rcmap, [])
        del self.save_destination

    def start_remote_load(self, destination):
        # This method is not reentrant. There is a race condition here.
        if hasattr(self, "load_destination"):
            raise RuntimeError("Concurrency is not allowed.")
        self.load_destination = destination
        rcmap = {}
        for s in self.segment_clients:
            rcmap[s] = s.run_command(f"startload {self.load_destination}")
        self._check_errors(rcmap, [])

    def load_handle(self, handle, registry_entry, typecode):
        if not hasattr(self, "load_destination"):
            raise RuntimeError("Can only call this when load is in progress.")
        new_handle = self._get_new_handle()
        self.variable_registry[new_handle] = registry_entry
        rcmap = {}
        for s in self.variable_registry[new_handle][2]:
            rcmap[s] = self.segment_clients[s].run_command(
                f"load {new_handle} {handle} {' '.join(typecode)}")
        self._check_errors(rcmap, [])
        return new_handle

    def finish_remote_load(self):
        if not hasattr(self, "load_destination"):
            raise RuntimeError("Can only call this when load is in progress.")
        rcmap = {}
        for s in self.segment_clients:
            rcmap[s] = s.run_command(f"finishload")
        self._check_errors(rcmap, [])
        del self.load_destination
        
    def clear_peer_var_store(self):
        # Runs the clear variable store command on each peer node
        def tmp(s):
            self.segment_clients[s].run_command("clearvariablestore 0")
        
        self.process_concurrently(tmp, [i for i in range(len(self.segment_clients))])

    def _cleanup(self, errmsg: List[str], new_handles: List[str]):
        if len(new_handles) > 0:
            rcmap = {}
            for s in range(len(self.segment_clients)):
                rcmap[s] = self.segment_clients[s].run_command("cleanup " +
                                                                " ".join(new_handles))
            for handle in new_handles:
                if handle in self.variable_registry:
                    del self.variable_registry[handle]
            cond = False
            for s in rcmap:
                parsed = rcmap[s].split(" ", 1)
                if parsed[0] == "error":
                    if not cond:
                        cond = True
                        errmsg.append("Errors detected in clean-up.")
                    errmsg.append(
                        f"{self.sm_name[s]} reporting error: {parsed[1]}")
        else:
            self._print("no handles to cleanup")

    def _check_errors(self, rcmap, new_handles):
        errmsg = []
        for s in rcmap:
            rc = rcmap[s]
            parsed = rc.split(" ", 1)
            if parsed[0] == "error":
                errmsg.append(
                    f"{self.sm_name[s]} reporting error: {parsed[1]}")
        if len(errmsg) != 0:
            self._cleanup(errmsg, new_handles)
            raise RuntimeError("ERROR: "+"\n".join(errmsg))
        for s in rcmap:
            if rc != rcmap[s]:
                errmsg = ["Unexpected type mismatch reported from segments."]
                self._cleanup(errmsg, new_handles)
                raise TypeError(errmsg[0])
        if rc == "ack":
            return ""
        return rc

    def process_command(self, request: str, scope: List[int]) -> dict:
        # Initiates running commands on peer nodes

        request = request.split(" ")
        num_new = int(request[1])
        del request[1]
        new_handles = [self._get_new_handle() for _ in range(num_new)]
        request[1:1] = new_handles
        rc = {}

        def tmp(s):
            rc[s] = self.segment_clients[s].run_command(" ".join(request))
        
        self.process_concurrently(tmp, scope)
        
        # The returned values include a copy of "new_handles". This is a
        # redundancy, and we currently don't check for errors in it.
        # This redundancy allows us to simply pass messages through, on their
        # way back from the segment managers.
        rc = self._check_errors(rc, new_handles)
        parsed = rc.split(" ")
        for i in range(0, 3 * num_new, 3):
            self.variable_registry[parsed[i + 2]
                                ] = (parsed[i], parsed[i + 1], scope)
            
        self.context().queue_for_deletion_clear()
            
        return rc

    def process_concurrently(self, f: Callable, iter: int):
        # Util method for running concurrent tasks
        tasks = []
        for i, v in enumerate(iter):
            tasks.append(threading.Thread(target=f, args=(v,)))
            tasks[i].start()
        for i in range(len(tasks)):
            tasks[i].join()