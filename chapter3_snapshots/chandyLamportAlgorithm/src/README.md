Chandy-Lamport algorithm works by having each component of the distributed system record its local state, as well as the messages it has received from other components, in a log. This log is then used to reconstruct the global state of the system. The algorithm proceeds as follows:

- Each component of the distributed system records its local state in a log.
- When a component receives a message from another component, it also records this message in its log.
- When a snapshot of the system is taken, each component sends its log to a central coordinator.
- The coordinator uses the logs from all the components to reconstruct the global state of the system at the time the snapshot was taken.