#!/bin/bash

if [ "$(id -u)" -ne 0 ]; then
        echo 'This script must be run by root' >&2
        exit 1
fi

read -p "Do you want to install the InstantMinecraft Server (y/n)?" choice
case "$choice" in
  y|Y ) echo "Installing...";;
  n|N ) exit;;
  * ) echo "invalid" && exit;;
esac

realuser="${SUDO_USER:-${USER}}"
arch=$(uname -m)

echo "Removing old installations..."
rm -rf /opt/instantminecraft

echo "Searching executable for architecture $arch..."

releasesJson=$(curl -s https://api.github.com/repos/InstantMinecraft/Server/releases/latest)
# Getting Tag name
tagName=$(echo $releasesJson | grep -o -P '(?<="tag_name": ").*(?=", "target_commitish)')
echo "Latest release is $tagName"

downloadUrl="https://github.com/InstantMinecraft/Server/releases/download/$tagName/instantminecraftserver_$arch"
path="/opt/instantminecraft/"
cd $path && cd ..
chown "$realuser" instantminecraft

echo "Downloading $downloadUrl to $path..."
filename="instantminecraft"
mkdir -p $path
wget "$downloadUrl" -O "$path$filename"
chmod +x "$path$filename"

rm /etc/systemd/system/instantminecraft.service
echo "Installing systemd service..."
echo "
[Unit]
Description=InstantMinecraft Server
ConditionPathExists=$path$filename
After=network.target
[Service]
User=$realuser
WorkingDirectory=$path
ExecStart=$path$filename
Restart=on-failure
RestartSec=5
[Install]
WantedBy=multi-user.target
" > /etc/systemd/system/instantminecraft.service

chown -R $realuser:$realuser /opt/instantminecraft

systemctl daemon-reload
systemctl enable instantminecraft.service
systemctl start instantminecraft.service
systemctl status instantminecraft
journalctl -u instantminecraft --no-pager

echo "Done!"