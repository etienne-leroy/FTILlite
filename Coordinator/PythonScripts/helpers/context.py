"""
Utility Helper Functions for FinTracer
"""

import ftillite as fl
import os
import sys
import logging
from logging.handlers import RotatingFileHandler


def create_logger(x, app_name):
    """Create a logger with file and stdout handlers"""
    logger = logging.getLogger(x)
    logger.setLevel(logging.INFO)
    formatter = logging.Formatter('%(asctime)s %(levelname)s: %(message)s')
    fn = f'logs/LOG-{x}-{app_name}.log'
    
    file_handler = RotatingFileHandler(fn, maxBytes=200000000, backupCount=5)
    file_handler.setFormatter(formatter)
    file_handler.setLevel(logging.INFO)
    logger.addHandler(file_handler)
    
    stdout_handler = logging.StreamHandler(sys.stdout)
    stdout_handler.setLevel(logging.WARNING)
    stdout_handler.setFormatter(formatter)
    logger.addHandler(stdout_handler)
    return logger

def setup_ftil_context(app_name):
    """Setup FTIL context with standard configuration"""
    # Ensure 'logs' directory exists
    if not os.path.exists("logs"):
        os.makedirs("logs")
    
    logger_all = create_logger('ALL', app_name)
    logger_client = create_logger('CLIENT', app_name)
    logger_compute_mgr = create_logger('COMPUTE MGR', app_name)
    logger_segment_client = create_logger('SEGMENT CLIENT', app_name)

    conf = fl.FTILConf().set_app_name("nonverbose") \
                        .set_rabbitmq_conf({'user': 'default_mq_user', 'password': 'default_mq_pw', 'host': 'localhost', \
                                            'COORDINATOR':'0', 'PEER_1':'1', 'PEER_2':'2', 'PEER_3':'3', 'PEER_4':'4'})\
                        .set_client_logger(logger_client)\
                        .set_compute_manager_logger(logger_compute_mgr) \
                        .set_segment_client_logger(logger_segment_client) \
                        .set_peer_names(['COORDINATOR', 'PEER_1', 'PEER_2', 'PEER_3', 'PEER_4']) \
                        .set_all_loggers(logger_all)
    
    return fl.FTILContext(conf=conf)


