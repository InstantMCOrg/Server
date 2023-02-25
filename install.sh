#!/bin/bash

if [ "$(id -u)" -ne 0 ]; then
        echo 'This script must be run by root' >&2
        exit 1
fi

read -p "Do you want to install the InstantMC Server (y/n)?" choice
case "$choice" in
  y|Y ) echo "Installing...";;
  n|N ) exit;;
  * ) echo "invalid" && exit;;
esac

realuser="${SUDO_USER:-${USER}}"
arch=$(uname -m)

filename="instantmc"
echo "Removing old installations..."
rm /opt/instantmc/$filename

echo "Searching executable for architecture $arch..."

releasesJson=$(curl -s https://api.github.com/repos/InstantMCOrg/Server/releases/latest)
# Getting Tag name
tagName=$(echo $releasesJson | grep -o -P '(?<="tag_name": ").*(?=", "target_commitish)')
echo "Latest release is $tagName"

downloadUrl="https://github.com/InstantMCOrg/Server/releases/download/$tagName/instantmcserver_$arch"
path="/opt/instantmc/"
cd $path && cd ..
chown "$realuser" instantmc

echo "Downloading $downloadUrl to $path..."
mkdir -p $path
wget "$downloadUrl" -O "$path$filename"
chmod +x "$path$filename"

frontEndReleasesJson=$(curl -s https://api.github.com/repos/InstantMCOrg/App/releases/latest)
# Getting Tag name
frontEndTagName=$(echo frontEndReleasesJson | grep -o -P '(?<="tag_name": ").*(?=", "target_commitish)')
frontEndUrl="https://github.com/InstantMCOrg/App/releases/download/$frontEndTagName/web.zip"

echo "Installing frontend files from $frontEndUrl"
mkdir -p "$path/frontend/"
wget "$frontEndUrl" -O "$path/frontend/web.zip"
cd "$path/frontend/"
unzip web.zip

rm /etc/systemd/system/instantmc.service
echo "Installing systemd service..."
echo "
[Unit]
Description=InstantMC Server
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
" > /etc/systemd/system/instantmc.service

chown -R $realuser:$realuser /opt/instantmc

systemctl daemon-reload
systemctl enable instantmc.service
systemctl start instantmc.service
systemctl status instantmc
journalctl -u instantmc --no-pager

echo "Done!"