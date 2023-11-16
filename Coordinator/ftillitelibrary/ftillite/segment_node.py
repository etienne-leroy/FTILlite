# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

class SegmentNode:

    def __init__(self, nodestring):
        nodeelements = nodestring.split(" ")
        self.nodeId = nodeelements[0]
        self.name = nodeelements[1]
        self.address = nodeelements[2]
        self.gpu = nodeelements[3]
    def __str__(self):
        return f"{self.nodeId}~{self.address}"
            