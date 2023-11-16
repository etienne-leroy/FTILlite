# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

import pika
import logging
import json
class SegmentClient:

    DEFAULT_HEARTBEAT_SECS = 600
    RETRY_ATTEMPTS = 3

    def __init__(self, compute_manager, name, num, rabbitmq_conf, logger):
        self.compute_manager = compute_manager
        self.name = name
        self.num = num
        self.variable_store = {}
        self.rabbitmq_conf = rabbitmq_conf
        self.logger = logger
        self.in_queue = f"{self.rabbitmq_conf.get('in_queue_prefix', 'FTILITE_INCOMING_')}{self.num}"
        self.out_queue = f"{self.rabbitmq_conf.get('out_queue_prefix', 'FTILITE_OUTGOING_')}{self.num}"
        self.corr_id = 0
        # Need to prevent pika from overloading log files
        logging.getLogger("pika").setLevel(logging.ERROR)
        self.connection = None

    def __del__(self):
        # Cleanup RabbitMQ Connection
        self.close_connection()

    def _print(self, value, loglevel=logging.INFO):
        if self.logger is not None:
            self.logger.log(loglevel, f"SEGMENT CLIENT ({self.name}, Node {self.num}), {value}")

    def run_command(self, request, response_required=True):
        self._print(f"COMMAND RECEIVED - {request}")
        self.corr_id += 1
        
        err = None
        for attempt in range(self.RETRY_ATTEMPTS):
            try:
                self.response = None
                self.channel = self.get_connection().channel()

                # Check in queue is open
                self.channel.queue_declare(queue=self.in_queue, passive=True)

                # Create out queue
                self.channel.queue_declare(self.out_queue, exclusive=True, auto_delete=True)

                self.channel.basic_consume(self.out_queue, self.on_response)
                
                MQ_properties = pika.BasicProperties(correlation_id=str(self.corr_id),
                                                     reply_to=self.out_queue if response_required else None)
                
                # Send to in queue
                self.channel.basic_publish(
                    exchange='',
                    routing_key=self.in_queue,
                    mandatory=True,
                    properties=MQ_properties,
                    body=json.dumps({
                            'command': f'command_{request}',
                            'response_required': str(response_required)
                    })
                )
                
                if not response_required:
                    self.get_connection().process_data_events()
                    self.channel.close()
                    return "ack"
            
                # Wait for data
                while self.response is None:
                    self.get_connection().process_data_events()
                    self.channel.queue_declare(queue=self.in_queue, passive=True)
                self.channel.close()
                break
            except (pika.exceptions.ConnectionClosed, pika.exceptions.StreamLostError) as ex:
                if attempt == self.RETRY_ATTEMPTS - 1:
                    err = f'Number of attempts exceeded. - {ex}'
                    break
                self._print(f'Queue connection is lost, attempting reconnection after error. - {ex}', logging.WARNING)
                self.close_connection()
            except pika.exceptions.UnroutableError as ex:
                err = f'Queue connection received no ACK. - {ex}'
                break
            except pika.exceptions.ChannelClosedByBroker as ex:
                err = f'Segment manager queue is down and the command has failed to process. The peer node may require restarting by an administrator. - {ex}'
                break
            except (pika.exceptions.ConnectionWrongStateError) as ex:
                err =  f"RabbitMQ connection is in a wrong state, attempting reconnection after error. - {ex}"
                self.close_connection()
        if err is not None:
            self._print(err, logging.ERROR)
            self.close_connection()
            return f"error {err}"
        return self.response

    def connect(self, heartbeat=DEFAULT_HEARTBEAT_SECS):
        try:
            param = pika.ConnectionParameters(
                host=self.rabbitmq_conf.get('host', 'localhost'),
                credentials=pika.PlainCredentials(username=self.rabbitmq_conf.get('user', 'guest'), 
                    password=self.rabbitmq_conf.get('password', 'guest')),
                heartbeat=heartbeat,
            )
            self.connection = pika.BlockingConnection(parameters=param)        
            self._print("RabbitMQ connection was successful.")
        except Exception as ex:
            self._print(f'RabbitMQ connection failed with error {ex}.', logging.ERROR)
            raise ex

    def on_response(self, ch, method, props, body):
        decoded_body = str(body.decode("utf-8"))
        if str(self.corr_id) != props.correlation_id:
            raise Exception(f"Response message from Segment {self.name} has different correlation ID. Expected: {self.corr_id}, actual: {props.correlation_id}")
        self._print(f"COMMAND RESPONSE - {decoded_body}")
        ch.basic_ack(method.delivery_tag)
        self.response = str(decoded_body)

    def get_connection(self):
        if not self.connection or self.connection.is_closed:
            self.connect()
        return self.connection

    def close_connection(self):
        try:
            self.connection.close()
        except Exception as ex:
            self._print(f"Attempted to close connection with exception: {ex}", logging.WARNING)
        self.connection = None

        