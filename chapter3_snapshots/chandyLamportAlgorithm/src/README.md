Chandy-Lamport algorithm works by having each component of the distributed system record its local state, as well as the messages it has received from other components, in a log. This log is then used to reconstruct the global state of the system. The algorithm proceeds as follows:

- Each component of the distributed system records its local state in a log.
- When a component receives a message from another component, it also records this message in its log.
- When a snapshot of the system is taken, each component sends its log to a central coordinator.
- The coordinator uses the logs from all the components to reconstruct the global state of the system at the time the snapshot was taken.


**ISSUES**

The Chandy-Lamport algorithm is a snapshot algorithm, which means that it takes a snapshot of the system state at a particular point in time and uses this information to determine the global state. This can cause some issues with the accuracy of the algorithm, because the global state that is reconstructed may not be the same as the global state at the time the snapshot was taken.

For example, if a message is sent from one component to another between the time the snapshot is taken and the time the logs are sent to the coordinator, this message will not be included in the global log and will not be part of the reconstructed global state. This can lead to inconsistencies and errors in the reconstructed global state, which can affect the accuracy and reliability of the algorithm.

Another issue with the Chandy-Lamport algorithm is that it can be computationally expensive, especially in systems with a large number of components. This is because each component has to record its local state and all the messages it receives in its log, and the coordinator has to reconstruct the global state from all these logs. This can require a significant amount of computation and storage, which can limit the scalability of the algorithm and make it less efficient for large systems.

Overall, the Chandy-Lamport algorithm is a simple and efficient algorithm for determining the global state of a distributed system, but it has some limitations and issues that need to be considered when implementing and using it.