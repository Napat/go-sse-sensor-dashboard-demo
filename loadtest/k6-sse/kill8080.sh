#!/bin/bash

lsof -ti:8080 | xargs kill -9