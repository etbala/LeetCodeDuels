"""
    Contains utility function to update upto which the problems has been downloaded, resetting configuration,
    reading upto which the problems has been downloaded
"""

def update_tracker(file_name, problem_num):
     """

     """
     with open(file_name, "w") as f:
         f.write(str(problem_num))

def reset_config():
    """
        Resets problem num downloaded upto to -1
        Resets  all the chapters
        Resets html file
    """
    update_tracker("track.conf", -1)

def read_tracker(file_name):
    """
    
    """
    with open(file_name, "r") as f:
        return int(f.readline())