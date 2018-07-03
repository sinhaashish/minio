#!/bin/sh
#
# Minio Cloud Storage, (C) 2017 Minio, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# If command starts with an option, prepend minio.
if [ "${1}" != "minio" ]; then
    if [ -n "${1}" ]; then
        set -- minio "$@"
    fi
fi

## Look for docker secrets in default documented location.
docker_secrets_env() {
    local ACCESS_KEY_FILE="/run/secrets/$MINIO_ACCESS_KEY_FILE"
    local SECRET_KEY_FILE="/run/secrets/$MINIO_SECRET_KEY_FILE"

    if [ -f $ACCESS_KEY_FILE -a -f $SECRET_KEY_FILE ]; then
        if [ -f $ACCESS_KEY_FILE ]; then
            export MINIO_ACCESS_KEY="$(cat "$ACCESS_KEY_FILE")"
        fi
        if [ -f $SECRET_KEY_FILE ]; then
            export MINIO_SECRET_KEY="$(cat "$SECRET_KEY_FILE")"
        fi
    fi
}

# switch to non root (minio) user if there isn't an existing .minio.sys directory or
# if its an empty data directory. We dont switch to minio user if there is an existing
# data directory, this is because we'll need to chown all the files -- which may take 
# too long, ruining the user's experience.
docker_switch_non_root() {
    for var in "$@"
    do
    echo " var $var"
     if [ "${var}" = "minio" ]; then
     continue
     fi
     if [ "${var}" = "gateway" ]; then
     owner="minio"
     continue
     fi    
     if [ "${var}" = "server" ]; then
     continue
     fi

    if [ -d "${var}" ]; then
            if [ -d "${var}/.minio.sys" ]; then
                owner=$(ls -ld "${var}/.minio.sys" | awk '{print $3}')       
            else
                # if the directory doesn't exist, this is a new deployment.
                mkdir -p "${var}"
                # change owner to minio user
                chown minio:minio "${var}"
                owner="minio"    
            fi
         break   
    fi
    done

    if [ "${owner}" = "minio"  ]; then
        # run as minio user
        exec su-exec $owner "$@"        
    elif [ "${owner}" = "root"  ]; then
        # else run as root user 
        exec "$@"
     else 
        echo " Exiting as user is not root or minio "
        exit 1
    fi
}

## Set access env from secrets if necessary.
docker_secrets_env

## Switch to non root user minio if applicable.
docker_switch_non_root "$@"