Tarry's algorithm is a algorithm for finding the shortest path between two nodes in a graph. It is a variant of Dijkstra's algorithm, which is a popular algorithm for finding the shortest path in a graph. While I am not able to browse the internet, I can provide a general outline for how to implement Tarry's algorithm in Go:

- Create a struct to represent a node in the graph. This struct should have fields for the node's value, its distance from the starting node, and a list of its neighbors and the weights of their edges.
- Create a priority queue to store the nodes in the graph. This queue should be sorted by the distance of each node from the starting node.
- Create a map to store the nodes that have been visited. This will be used to avoid visiting the same node more than once.
- Add the starting node to the priority queue and set its distance to 0.
- While the priority queue is not empty, do the following:
    - Remove the node with the smallest distance from the priority queue.
    - If the node has not been visited, mark it as visited and add its neighbors to the priority queue with their distances updated according to the edge weights.
- Once the algorithm has finished, the shortest path to the destination node can be reconstructed by following the "previous" pointers of each node.