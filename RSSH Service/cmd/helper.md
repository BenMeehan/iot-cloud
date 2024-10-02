ssh -p 2222 ghost@localhost

ssh -p 2222 user@localhost "ls -al"

PF:
ssh -L 8080:localhost:8080 -p 2222 user@localhost

RSSHL
ssh -R 8080:localhost:8080 ghost@10.0.2.15 -p 2000