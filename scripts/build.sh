#!/bin/bash

# build frontend
cd front
npm install
npm run build

# build backend
cd ..
go build -o app .