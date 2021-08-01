# Porter In-Cluster Agent

This repository contains the source code for an in-cluster agent that aggregates cluster events and forwards them to a Porter server. This agent forwards three types of events:
- Pod-level events, for example if a Pod terminates due to an application-level error. Ideally, the agent is able to keep a buffer of the previous 100 log lines from `stderr` and `stdout` for each Pod, and will bundle the termination event with those logs, so that application developers can view those logs later on. 
- HPA-level events, which shows reasons for autoscaling. DevOps engineers (and sometimes application developers) should be able to view the reasons for horizontal pod autoscaling triggers, such as CPU/memory/custom metric. 
- Node-level events, in particular when a node is unhealthy. 

This agent forms the basis for an events tab on the Porter dashboard, along with notifications for users when deployments/apps scale, restart, or when machines terminate.  
