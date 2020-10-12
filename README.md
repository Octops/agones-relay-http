# Agones Relay HTTP

Notify HTTP endpoints when state changes events happen to GameServers or Fleets. 

The published payload contains an entire representation of the GameServer or Fleet at the moment when event the got fired.

The possible types of events are OnAdd, OnUpdate and OnDelete.

**When a GameServer or Fleet state change event gets fired?**

A few examples are:
- The state of the GameServer changed during its all lifespan. Scheduled, Ready, Allocated, Shutdown, etc. 
- The GameServer status fields like address, port or player tracking fields.
- Then the number of replicas of a Fleet went up or down.
- Fleets status fields like: players capacity or count; allocated replicas, ready replicas and reserved replicas.
 
In addition to the state change events, the Agones Relay HTTP reacts to `reconcile` events. On an interval bases a complete state of the world gets published. That means all the current GameServers and Fleets states.
The reconcile interval is controlled by the flag `--sync-period`. Make sure you don't set this value to low. The consequence could be a DDoS against the endpoints.

**Important**

OnAdd, OnUpdate and OnDelete are not controled by the reconcile interval. Those events get published at the time they happen.  

## Verbs and Payloads 
| Event        | Verb          | Payload             |
| ------------ |:------------- | ------------------: |
| OnAdd        | POST          | New Obj             |
| OnUpdate     | PUT           | Old and New Obj     |
| OnDelete     | DELETE        | None (Query Params) |

Examples of the POST and PUT payloads can be found on [examples/payloads]() folder.

For `OnDelete` events the body of the request is null. The URL of the request looks similar to:
```
DELETE http://localhost:8090/webhook?event_type=gameserver.events.deleted&name=simple-udp-pmx5c-xzfft&namespace=default
DELETE http://localhost:8090/webhook?event_type=fleet.events.deleted&name=simple-udp&namespace=default
```
List of Params

| Param        | Description               | Example                                         | 
| -----------  |:------------------------- | ----------------------------------------------- |
| name         | name of the resource      | simple-udp-pmx5c-xzfft                          |
| namespace    | namespace of the resource | default                                         |
| event_type   | type of the event fired   | fleet.events.deleted, gameserver.events.deleted |

## Diagram?

What kind of payload and Verbs

## How to Install

Update the `install.yaml` file to reflect the endpoints that must be notified.

The list of endpoints can be specific for each kind of event if that is a requirement. Otherwise, just use the `--on-event-url` flag.

The url flags can be a list separated by comma.
```
--on-add-url=http://www.myendpoint.com,http://www.anotherendpoint.com/webhooks
``` 

```yaml
args:
    - --sync-period=15s # period between every reconcile cycle 
#   - --on-add-url=http://www.myendpoint.com/onadd
#   - --on-update-url=http://www.myendpoint.com/onupdate
#   - --on-delete-url=http://www.myendpoint.com/ondelete
    - --on-event-url=http://www.myendpoint.com/webhook
    - --verbose
```

Push the manifest that will create the required service account, RBAC and deployment.
```bash
$ kubectl -f deploy/install.yaml
```

## Local Server
```bash
# Use the flag  --verbose=true to output the whole received request body
$ go run hack/server.go --verbose=false --addr=":8090" 
http server started on [::]:8090
INFO[0004] webhook received: onupdate/gameserver.events.updated
INFO[0004] webhook received: onupdate/gameserver.events.updated
INFO[0004] webhook received: onupdate/fleet.events.updated
INFO[0024] webhook received: onupdate/fleet.events.updated
INFO[0024] webhook received: onadd/gameserver.events.added
INFO[0024] webhook received: onadd/gameserver.events.added
INFO[0024] webhook received: onadd/gameserver.events.added
INFO[0024] webhook received: onadd/gameserver.events.added
INFO[0024] webhook received: onadd/gameserver.events.added
INFO[0024] webhook received: onupdate/gameserver.events.updated
INFO[0024] webhook received: onupdate/gameserver.events.updated
INFO[0024] webhook received: onadd/gameserver.events.added
INFO[0024] webhook received: onupdate/gameserver.events.updated
INFO[0024] webhook received: onupdate/gameserver.events.updated
INFO[0024] webhook received: onupdate/gameserver.events.updated
INFO[0083] webhook received: ondelete/gameserver.events.deleted default-simple-udp-9nt4c-qv7nm
INFO[0087] webhook received: ondelete/gameserver.events.deleted default-simple-udp-9nt4c-t46nx
INFO[0089] webhook received: ondelete/gameserver.events.deleted default-simple-udp-9nt4c-pt5fb
INFO[0150] webhook received: ondelete/fleet.events.deleted default-simple-udp
```