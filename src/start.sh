#!/bin/bash

ps aux | grep rate-limit | grep -v grep | grep -v Visual | awk '{print $2}' | xargs kill -9

nohup ./rate-limit &
