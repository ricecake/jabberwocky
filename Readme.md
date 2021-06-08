Jabberwocky!
Goals:
No need for load balancer, servers tell clients where to connect.
All custom scripts are signed
All executed programs are on whitelist
Actions are logged
Agent has a persistent connection to server
Server pushes a message down, and results are attached to the same envelope.  Agent will preserve any metadata attached to command.
Agents will prefer to use dns to get server list, but can also use a seed node - reroute happens either way using lrw hashing
lrw hash weight should be based on the memberlist 'GetHealthScore' funciton.  Along the lines of 1/max(1, 1+GetHealthScore), with a minimum of 1, and then gcd with the lowest value being pegged to 1.
sqlite for local database

Server needs to ability to serve static files
that way there are commands that can be given around pulling down files.
It should also be able to dynamically fetch files from other places, or to direct the client to pull from arbitrary urls.
Seems a bit much to have it push the file over the websocket.
Pull from url, defaulting to self hosted, makes it a bit easier to cdn and be sensible.

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

should have an option to do ssdp clustering, but also make a client command to tell a node to join a seed node, or to configure a seed node in the config file


Need to add a message type that agents can emit that sets tags.  The tags can get set on their connection in the server, and used to send them messages.
So need: output, log, alive, and setTag
These should respect fields they've been given when the job was installed, to allow for 'reply-to' style functionality.  So stored jobs on agent need to be able to keep job metadata
should also include time fields and the like.  duration, and errors.  start finish and job id. jobs should be able to accept an id during installation, or make one up if they don't get one.

Server needs the notion of different output backends.
Server should support file, amqp, mqtt, nsq and http backends - should always have admin ws tap option
Should have the ability to run a script to format output before sending to backend
Should have the ability to accept a tag for the admin websocket connection, to filter what messages get sent over websocket.
Should gossip what servers are in cluster, and what agents, and what agents live on which server.
If trying to examine log output from an agent not on this server, should be able to link to the appropriate server, rather than try to stream logs from one to the other.
Maybe can just connect the websocket to that server?  Might be easier.

Need the ability for the server to run a script to check connection authentication.  Server scripts might need to have a couple extra commands available to them to do this.  Or just allow auth via http call for initial key negotiation?  Prefer script, with http capability.  That way can call out to an auth server which can approve based on tags associated with connecition, like IP address.


When starting for the first time, and agent should create a public key, and send it to the server.  The server should store and gossip this key.
When getting a connection, the server should send a challange message to the agent, which consists of a timestamped nonce and keyid, and the agent should a challange response
that is the jwt for signing the given body.  Basically the server sends a jwt header and body, and the agent signs it, and the server checks that it all lines up.

Should see if it's possible to support a plugin architecture for the js script functions.  That way downstream users can compile their own, without needing to modify the source.

Should, if supporting basic auth, use something like bcrypt to store password hashes in config file, rather than raw passwords or sha based.


Should flesh out the different type of commands more.
Need at basic: script, execute a command, watch for changes in a directory, and tail a log file.
Should also have "watch journalctl for a unit", which will handle the muck of json encoding it and whatnot.
Should also expose that as a js builtin, so that scripts can filter.

Should basically all commands be js functions?  Do they need to be standalone, if a simple js script can execute?
Would be nice for simplicity.  But should see what can be done for code sharing, so that the command can send the same args that the js wrapper would expect.
Have each command expose a category, and a name.  Then install in in js so that it's called as "category.name()"

in addition to text log functions, error, warn, info, debug, trace, need a way to emit a structured object as the output.  Shouldn't be too fancy, just a way to say "there's text", and "there's structures".  Don't support top level arrays, but allow them lower in the object.


Need to develop an intermediary delegate for server command execution.
giving a command via tool, web should gossip the command outwards to other nodes.
Should have a delegate that handles events that come in via different channels, so that it's consistent in how it handles the events, regardless of their origin.

Should also move the different commands into a seperate file.
There should be a top level "command" management file under the payload module, and then different files for different  ppaayload categories.
Larger categories may have multiple files, but it's entirely predicated on what makes it easier to find commands.

Commands should have a "key", which is a dot delimited list of nested categories, with the final label being the command name.
might split that into a dot command string, and a seperate name string, but probably not.

Should use afero for a mock filesystem that can be handled by tests/reality.
Should make an interface for a "command executor" that will typically just be a pass through, probably with logging and such.
Should also be a system for enforcing command acls, which might be returned from the executor delegate.
Goal is to make things mockable for better testing.

Should consider the scope of whats handled by the gossip delegate.  Should it be all gossip messages, or just broadcast commands?
Leaning all gossip events, since it makes sense to handle node join/remove via the same mechanism as a broadcast signal about a client joing a server.
Thinking is required.


The payload execute method should be more focused on figuring out what type of job it is, and creating as needed, than on immediate execution.
Need a Job db type, which if it's a persistent job can be recorded.
the basic flow should be "create job -> record if persistent -> initiate job".  Depending on type, initiation may be direct execution, or it might be cron registration, etc, etc.
When starting the system, it will do setup, and then initiate jobs based on what time period different jobs need to be run.
on start, on connect, interval, cron, on stop, on disconnect

registering a script should be a different command than executing the script.  That way scripts are seen as generally persistent things.  scripts will also need to be deleted.

need to filter based on type, with the "command" type being what precipitates the creation and initiation of job records.


for now, keep output options simple.  Always support a "firehose" websocket.  If anyone connects to it, they'll get the stream.
Also include AMQP support, since I want to use that.  In the future, consider nsq, and outbound websocket stream


Should make the gossip parts gossip signed messages, to protect against unintended joints.  Basically just use hmac signed preshared keys, in a jwt setup.
Not huge priority, but should be easy to do.

Next priority: Work on payload/command installation and job creation. Figuring out auth model can wait a little bit, once core functionality is finished. 
