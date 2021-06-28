# How to deploy me


- access the server (request ip and username from me)
- Use systemd script
- request access to the repo (https://github.com/noebs/ondemand)
- Download go (https://golang.org/dl) if not available
- Setup Gopath and all the necessary steps (also available at the https://golang.org go read instructions)
- `git clone` the repository
- `go build` (in the root dir)
    - it might take some time to get all dependencies and so
- add a new server block to nginx for this api endpoint (i can help with this)
- reload nginx 


You can use this as an example for systemd unit services

```systemd
[Unit]
Description=Ondemand
ConditionPathExists=/home/ubuntu/src/ondemand
After=network.target
 
[Service]
Type=simple
User=ubuntu
Group=ubuntu
LimitNOFILE=1024

Restart=on-failure
RestartSec=10
startLimitIntervalSec=60

WorkingDirectory=/home/ubuntu/src/ondemand
ExecStart=/home/ubuntu/src/ondemand/ondemand

# make sure log directory exists and owned by syslog
PermissionsStartOnly=true
ExecStartPre=/usr/bin/mkdir -p /var/log/ondemand
ExecStartPre=/usr/bin/chown syslog:adm /var/log/ondemand
ExecStartPre=/usr/bin/chmod 755 /var/log/ondemand
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=sleepservice
 
[Install]
WantedBy=multi-user.target

```
