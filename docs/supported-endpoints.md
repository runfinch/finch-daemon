## Supported API Endpoints


The finch-daemon focuses on core container orchestration functionality and omits advanced features like Swarm mode, plugins, and many administrative/maintenance operations. [Unsupported APIs](#unsupported-apis)

> **Note**: This document lists API endpoints that are functionally available in finch-daemon. However, implementations are provided on a best-effort basis and may not support all query parameters, request options, or edge cases present in the official Docker API v1.43 specification. Behavioral differences may exist due to the underlying containerd/nerdctl implementation. Users should thoroughly test their workflows and consult the source code for specific implementation details.


### System APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/info` | GET | Get system information |
| `/version` | GET | Get version information |
| `/_ping` | HEAD, GET | Ping the daemon |
| `/auth` | POST | Check auth configuration |
| `/events` | GET | Monitor events |

### Container APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/containers/json` | GET | List containers |
| `/containers/create` | POST | Create a container |
| `/containers/{id}/json` | GET | Inspect a container |
| `/containers/{id}/start` | POST | Start a container |
| `/containers/{id}/stop` | POST | Stop a container |
| `/containers/{id}/restart` | POST | Restart a container |
| `/containers/{id}/kill` | POST | Kill a container |
| `/containers/{id}/pause` | POST | Pause a container |
| `/containers/{id}/unpause` | POST | Unpause a container |
| `/containers/{id}/wait` | POST | Wait for a container |
| `/containers/{id}/remove` | POST | Remove a container |
| `/containers/{id}` | DELETE | Remove a container (alternative) |
| `/containers/{id}/attach` | POST | Attach to a container |
| `/containers/{id}/logs` | GET | Get container logs |
| `/containers/{id}/stats` | GET | Get container stats |
| `/containers/{id}/top` | GET | List processes running inside a container |
| `/containers/{id}/rename` | POST | Rename a container |
| `/containers/{id}/exec` | POST | Create an exec instance |
| `/containers/{id}/archive` | GET | Get an archive of files/folders |
| `/containers/{id}/archive` | PUT | Extract an archive to a directory |

### Image APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/images/json` | GET | List images |
| `/images/create` | POST | Pull an image |
| `/images/load` | POST | Load a tarred repository |
| `/images/{name}/json` | GET | Inspect an image |
| `/images/{name}/push` | POST | Push an image |
| `/images/{name}/tag` | POST | Tag an image |
| `/images/{name}` | DELETE | Remove an image |
| `/images/{name}/get` | GET | Export an image |

### Network APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/networks/` | GET | List networks |
| `/networks` | GET | List networks (alternative) |
| `/networks/create` | POST | Create a network |
| `/networks/{id}` | GET | Inspect a network |
| `/networks/{id}` | DELETE | Remove a network |
| `/networks/{id}/connect` | POST | Connect a container to a network |

### Volume APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/volumes` | GET | List volumes |
| `/volumes/create` | POST | Create a volume |
| `/volumes/{name}` | GET | Inspect a volume |
| `/volumes/{name}` | DELETE | Remove a volume |

### Exec APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/exec/{id}/start` | POST | Start an exec instance |
| `/exec/{id}/resize` | POST | Resize an exec instance |
| `/exec/{id}/json` | GET | Inspect an exec instance |

### Distribution APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/distribution/{name}/json` | GET | Get image information from registry |

### Build APIs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/build` | POST | Build an image from a Dockerfile |

## Unsupported APIs

#### Swarm APIs
- **Services**: `/services/*` - Manage services in a swarm
- **Nodes**: `/nodes/*` - Manage nodes in a swarm
- **Swarm**: `/swarm/*` - Manage swarm
- **Tasks**: `/tasks/*` - List tasks
- **Secrets**: `/secrets/*` - Manage secrets
- **Configs**: `/configs/*` - Manage configs

#### Plugin APIs
- **Plugins**: `/plugins/*` - Manage plugins

#### Session APIs
- **Session**: `/session` - Initialize interactive session

