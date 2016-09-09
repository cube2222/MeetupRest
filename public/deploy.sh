#!/bin/bash
##Remove node_modules folder
npm install
npm run build
rm -rf node_modules
goapp deploy ../.
npm install
