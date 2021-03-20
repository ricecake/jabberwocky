Jabberwocky!
Goals:
No need for load balancer, servers tell clients where to connect.
All custom scripts are signed
All executed programs are on whitelist
Actions are logged
Agent has a persistent connection to server
Server pushes a message down, and results are attached to the same envelope.  Agent will preserve any metadata attached to command.
Agents will prefer to use dns to get server list, but can also use a seed node - reroute happens either way using lrw hashing
sqlite for local database

commands:
server - runs the server
agent - runs the agent
sign - used to sign a custom script, and upload to the cluster
client - used to query the server cluster

Authentication:
 provide a set of options
 fixed credentials - creds in config
 proxy auth - if the proxy says you're okay...
 url based - pass the credentials to a url as http basic, and if get 200, it's good
 openid?
   would need to be traditional flow, not spa
   only a handful of endpoints to support
   stretch
    

 basic notion is that we should be able to have the system accept a handful of auth options, and then do the right thing.
 the ui is basically dumb
