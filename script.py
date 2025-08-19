#!/usr/bin/env python3
"""
Script
"""

import os
import subprocess
import time


def main():
    """
    main
    """
    file_name = "index.html"
    last_save_time = os.path.getmtime(file_name)

    print(last_save_time)


if __name__ == "__main__":
    main()
