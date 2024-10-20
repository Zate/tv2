#!/bin/bash
TALOS_VER=1.7.4
CLUSTER_NODES=(3)
CLUSTER_NAME=taipan
CLUSTER_IP=192.168.0.10
#NODE_3_NAME=node3.zate.systems
#NODE_3_IP=192.168.0.13
FILE=/mnt/sdcard/metal-${TALOS_VER}-arm64.raw
# /usr/bin/printf "[ \xE2\x9C\x94 ] \t${FILE} exists\n"
[[ ! -f ${FILE} ]] && { /usr/bin/printf "[ x ] \t${FILE} doesnt exist"; curl -Lks https://github.com/nberlee/talos/releases/download/v${TALOS_VER}/metal-arm64.raw.xz -o ${FILE}.xz; unxz ${FILE}.xz; } || { /usr/bin/printf "[ \xE2\x9C\x94 ] \t{FILE}exists\n"; }
for i in ${CLUSTER_NODES[@]}; do
        /usr/bin/printf "[ = ] \tinstalling Talos ${TALOS_VER} on node ${i}\n"
        echo "tpi flash --local --image-path ${FILE} --node ${i}"
        tpi flash --local --image-path ${FILE} --node ${i}
        [[ $? -eq 0 ]] && { /usr/bin/printf "[ \xE2\x9C\x94 ] \tFlash Complete\n"; } || { /usr/bin/printf "[ x ] \tFlash Failed\n"; exit 1; }
        tpi power on --node ${i}
        /usr/bin/printf "[ \xE2\x9C\x94 ] \tnode ${i} powered on\n"
        for s in 5 4 3 2 1; do
                sleep 1
                /usr/bin/printf "[ ${s} ]\n"
                tpi uart --node ${i} get | tee >> ./uart.${i}.log
        done
        mkdir -p /mnt/sdcard/${CLUSTER_NAME}
        cd ${CLUSTER_NAME}
        /mnt/sdcard/talosctl gen config ${CLUSTER_NAME} https://192.168.0.1{i}:6443 --install-disk /dev/mmcblk0
        cd ..
done

# curl -LOk https://dl.k8s.io/v1.30.2/bin/linux/arm/kubectl

# /mnt/sdcard/talosctl gen config taipan https://192.168.0.13:6443 --install-disk /dev/mmcblk0
# - this does the install to the node, and installs to the mmc.
# - unknown how we ensure that the nvme is available for use for storage at some point.
# vi controlplane.yaml and worker.yaml
# - find the line starting with # kernel:
# - add in lines to load the rockchip-cpufreq kernel module
# - likely also where we add in any other modules we added.
# /mnt/sdcard/talosctl apply-config --insecure -n 192.168.0.13 -f controlplane.yaml
# /mnt/sdcard/talosctl config merge /mnt/sdcard/taipan/talosconfig
# - maybe we can edit the talosconfig to add in the endpoints first?
# vi /root/.talos/config
#  - edit the line with endpoints: [] to be a yaml array with the node ip
# /mnt/sdcard/talosctl get extensions -n 192.168.0.13
# /mnt/sdcard/talosctl read /proc/cpuinfo -n 192.168.0.13
# /mnt/sdcard/talosctl read /proc/modules -n 192.168.0.13
# /mnt/sdcard/talosctl dashboard -n 192.168.0.13
# /mnt/sdcard/talosctl bootstrap -n 192.168.0.13