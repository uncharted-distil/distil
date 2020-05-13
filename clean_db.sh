#!/bin/bash
sudo rm -rf $D3MOUTPUTDIR
docker system prune --force && docker volume prune --force
