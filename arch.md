### Components
The proposed components are roughly defined below:
1. We have the main `ControllerManager` deployed as a k8s controller
2. `ControllerManager` encapsulates multiple `controllers` each listening to the following resource specific events:
    * `Pod`
    * `HPA`
    * `Node`
    > Each controller listens to just a single type of resource events
3. A `Redis` server holding a queue of logs against each individual resource like `Pod` (with namespace-named context)
4. A filter list for the kind of events which needs to trigger a notification event
5. An HTTP client residing inside the controller itself to push to porter server
6. A `Configmap` for configuring the porter server location, credentials, ports etc.

### Behaviour
1. For every event the controller receives, it runs it against a filter list to determine if its a critical event that needs a notification.
    * If not, the controller will just get the pod's logs and push it to the relevant queue in redis truncating at max 100 lines.
    * If it is a critical event, it fetches the latest logs, merges with the existing entries in redis and pushes to the porter
based on the config provided in the config map.
2. The following is the proposed data model for the redis store – 
    * Approach 1: A key for each resource with the following convention `<resource_type:Namespace:Name>` containing a `sorted set` of logs trimmed at 100.
    The structure would roughly look like this:
    ```
    "pod:default:my-pod-abc-xyz" => {"log entries as strings", "not more that 100", "in length"},
    "hpa:default:my-hpa-" => {"log entries as strings", "not more that 100", "in length"}
    ```
    * Approach 2: Separate `redis namespaces` for each resource type. Rest all remains the same.
3. Regarding the Agent-to-Server communication for posting the notification, following can be a basic starter for the JSON payload:
    ```json
    {
      "resource_type": "Pod",
      "name": "pod_name",
      "namespace": "namespace_name",
      "cluster": "cluster_name",
      "message": "<message>",
      "reason": "<reason>",
      "tail_logs": [
        "last 100 log lines",
        "as an array of strings",
        "..."
      ]
    }
    ```

### Design choices
1. Since the porter server is currently HTTP and the communication between controller and server is not going to be an overwhelming number
of requests, we have decided to stick with HTTP & JSON based communication for now over a more advanced use like gRPC.


### Diagram
![porter-agent](https://user-images.githubusercontent.com/7482025/127897279-6be8a8bd-8dfc-40c9-b103-33085d87582f.png)

