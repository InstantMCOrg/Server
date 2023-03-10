# _InstantMC/_**Server**

🚀 An unbelievable fast Minecraft Server management tool \
⚡ Minecraft Server ready in **under one second** \
🔐 Secure individual management

[![Test](https://github.com/InstantMCOrg/Server/actions/workflows/test.yaml/badge.svg)](https://github.com/InstantMinecraft/Server/actions/workflows/test.yaml)

# How does it work?
Your Minecraft Server run in dedicated docker container with a [custom runtime and backend](https://github.com/InstantMC/Client).
The InstantMC Server prepares them (world generation etc.) in the background and pauses the docker container. If you want to start a Minecraft server, this backend searches for available prepared Minecraft server with the target requirements. If a prepared unit has been found it will be resumed, and it will be up and running in under one second.
If no prepared server fits the requirements a new one will be generated and activated as soon as the Minecraft world is ready.

# Installation
## Prerequisites
- CPU architecture must be either x86_64 or arm64
- Docker is installed and active
- Your system is using systemd
- You opened the ports 25000-25090 (the http server listens on port 25000)
## Install and run the software using the ``install.sh`` script:
```bash
$ wget https://raw.githubusercontent.com/InstantMCOrg/Server/main/install.sh -O install.sh
$ chmod +x install.sh
$ sudo ./install.sh
```

# Usage
## Using the HTTP-API
_The HTTP server is listening on port 25000_

`GET /api/` \
Response:
````json
{
  "server": "InstantMC"
}
````

### Note: First of all, you need to login and obtain your session
First-Time-Login after installation:

`POST /api/login` \
_Form Body:_ \
`username : admin`\
`password : admin`

This will be the response:
```json
{
  "token": "<YOUR-TOKEN>",
  "password_change_required": true
}
```
_You'll get your token to use the HTTP api, but you need to change the admin password. Do the following:_

`POST /api/user/password/change` \
_Headers:_ \
`auth : <YOUR-TOKEN>`\
_Form Body:_ \
`password : <YOUR-NEW-PASSWORD>`

This will be the response:
```json
{
  "token": "<YOUR-TOKEN>"
}
```
_The old token won't work any more after you changed your password. Use the new one._

### After you completed the _First-Time-Login_-steps you can access the REST HTTP API:
**Note: you need to send the following header in every request:** `auth: <YOUR-AUTH-TOKEN>`

`GET /api/server` \
Response example:
````json
{
    "server": [
        {
            "server_id": "bead864ef09219d6aa29d2702204f90d",
            "name": "My totally no cheats world",
            "mc_version": "1.19.3",
            "port": 25056,
            "ram_size_mb": 1024,
            "status": "Running"
        }
    ]
}
````


`GET /api/server/prepared` \
Response example:
````json
{
  "prepared_server": [
    {
      "number": 0,
      "mc_version": "1.19.3",
      "ram_size_mb": 1024
    },
    {
      "number": 1,
      "mc_version": "1.19",
      "ram_size_mb": 2048
    }
  ]
}
````

`POST /api/server/start` \
_Form values:_
```
name: <YOUR-NAME>
mc_version: 1.19.3
ram: 1024
```
_Note: A list of available mc-versions can be found [here](https://github.com/InstantMC/Server/blob/faab69f5ca42bb4d7dec472e0e42a9eeca7f1724/pkg/config/mccontainer.go#L16)_ \
_Note: RAM size is in mb and is optional (1024 is default)_

Response example: \
_If a prepared server has been picked up and started instantly_
````json
{
  "server_id": "b29a482b685d7bcb683b73fc2bf76bcd",
  "name": "My world",
  "mc_version": "1.19.3",
  "port": 25042,
  "ram_size_mb": 1024,
  "status": "Running"
}
````
_If a server needs to be prepared before start_
````json
{
  "mc_version": "1.19.3",
  "name": "Meine Minecraft Weld",
  "ram_size_mb": 1024,
  "server_id": "bead864ef09219d6aa29d2702204f90d",
  "status": "Preparing"
}
````
_In that case take a look at the `ws /server/start/status/` request_


`DELETE /api/server/<SERVER-ID>/delete` \
Response example:
````json
{}
````

`ws /api/server/start/status/<SERVER-ID>` \
_Retrieves status about the server preparation_ \
Message Examples:
````json
{
  "message": "Preparing server preparation"
}
````
````json
{
  "message": "Starting server preparation"
}
````
````json
{
  "message": "Preparing world 2%"
}
````
````json
{
  "message": "Preparing world 99%"
}
````
````json
{
  "message": "Preparing world 100%"
}
````
````json
{
  "message": "Done"
}
````

**More APIs to be added soon**


## Using the dedicated Flutter App
_Coming soon_