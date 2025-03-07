# Custom SSH Service 

## Security Considerations 

This custom SSH service is designed for testing purposes only and should not be used in a production environment.

1. Password-based authentication for the root user is enabled, which is generally not recommended for security reasons.
2. Key-based authentication authentication is crucial for securing your SSH service. Below are the steps to generate SSH keys. 


## Configure Password-Based Authetication
Inside the directory where  `Dockerfile` for the SSH service is in, create a file named `ssh_users.txt` . In this file, add your FTP user credentials in the format `username:password`.

Example `ssh_users.txt`:
```
root:password
```


## Generating SSH Keys 

Generate keys in `service/custom_sshd/keys`:

```
$ ssh-keygen -t rsa -b 4096 -C "riotpot@riotpot.com" -f ./keys/riopot_rsa
```

or 

```
$ ssh-keygen -C "riotpot@riotpot.com" -t ed25519 -f ./keys/riotpot_ed25519
```

