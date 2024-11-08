# Consistent Hashing Project

This project implements a consistent hashing algorithm with a server pool for load balancing. The consistent hashing algorithm ensures that the distribution of keys to nodes is balanced and minimizes the number of keys that need to be remapped when nodes are added or removed.

## Features

- **Consistent Hashing**: Efficiently distributes keys across nodes.
- **Server Pool Management**: Add and remove nodes from the server pool.
- **Jump Hash**: Implementation of the Jump Hash algorithm for consistent hashing.
- **Memento Hash**: Implementation of the Memento Hash algorithm for consistent hashing.

## Project Structure

- `loadbalance.go`: Entry point of the application.
- `servernode.go`: Implementation of a simple server node.
- `consistenthash`: Implementation of a generic conistent hasher
- `hashing/`: Package for hashing utilities.
- `serverpool/`: Package for managing the server pool.