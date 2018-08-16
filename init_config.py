# -*- coding:utf-8 -*-

import os
import shutil

def main():
    for root, _, files in os.walk(os.curdir):
        for filename in files:
            name, ext = os.path.splitext(filename)
            if ext != '.example':
                continue
            newfile = root + os.path.sep + name
            fullpath = root + os.path.sep + filename
            shutil.copyfile(fullpath, newfile)

if __name__ == '__main__':
    main()
