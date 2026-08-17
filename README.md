[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![GPL License][license-shield]][license-url]
[![Build Status][build-shield]][build-url]


<div align="center">
  <a href="https://mono.tr/">
    <img src="https://r2.mono.tr/logo/Mono-Logo.svg" width="340"/>
  </a>

  <h1 align="center">vgw-manager</h1>
</div>

`vgw-manager` is a comprehensive management and provisioning tool for VersityGW, bridging the gap between ZFS dataset management and VersityGW's admin APIs. It allows for seamless user and bucket administration with integrated quota enforcement.

## Features

*   **User Management**: Create, update, and delete users in VersityGW.
*   **Bucket Management**:
    *   Create and delete buckets with ZFS backend integration.
    *   Enforce storage quotas at the filesystem level.
    *   Manage bucket ownership and Access Control Lists (ACLs).
    *   Toggle bucket visibility (Public/Private).
*   **Provisioning**: A single-command provisioning workflow to set up a user and their primary bucket instantly.
*   **Interactive TUI**: A rich, easy-to-use Terminal User Interface for interactive management.
*   **CLI Interface**: Full non-interactive command-line support for automation and scripting.

## Installation

### Prerequisites

*   Go 1.22+
*   ZFS (zfs-utils/zfs-fuse) installed and configured on the host.
*   VersityGW running with Admin API enabled.

### Download Binary (Recommended)
Download the latest binary for your operating system from the [Releases](https://github.com/monobilisim/vgw-manager/releases) page.

```bash
# Example for Linux amd64
wget https://github.com/monobilisim/vgw-manager/releases/latest/download/vgw-manager_Linux_x86_64.tar.gz
tar xvf vgw-manager_Linux_x86_64.tar.gz
sudo mv vgw-manager /usr/local/bin/
```

### Build from Source


```bash
git clone https://github.com/monobilisim/vgw-manager.git
cd vgw-manager
make build
```

This will produce the `vgw-manager` binary in the current directory.

To install system-wide:
```bash
sudo make install
```

## Configuration

The application authenticates with VersityGW and manages ZFS datasets. Configuration can be provided via a YAML file or environment variables.

### Config File
Default location: `/etc/vgw-manager.yaml`

An example configuration file is provided in the repository: [`vgw-manager.example.yaml`](vgw-manager.example.yaml).

```yaml
# vgw-manager configuration example
# Copy this file to /etc/vgw-manager.yaml or pass via --config flag

# VersityGW Admin Credentials
adminAccess: "changeme-access"
adminSecret: "changeme-secret"

# VersityGW Endpoint
endpointURL: "http://localhost:7070"
region: "us-east-1"

# Paths
usersJSONPath: "/tank/s3/accounts/users.json"
zfsPoolBase: "tank/s3/buckets"
mountBase: "/tank/s3/buckets"
```



### Environment Variables

| Variable | Description |
|----------|-------------|
| `VGW_ADMIN_ACCESS` | VersityGW Admin Access Key |
| `VGW_ADMIN_SECRET` | VersityGW Admin Secret Key |
| `VGW_ENDPOINT_URL` | VersityGW Endpoint URL |
| `VGW_ZFS_POOL_BASE` | Base ZFS pool/dataset for buckets (e.g., `tank/s3`) |
| `VGW_USERS_JSON_PATH` | Path to `users.json` for read operations |
| `VGW_API_LISTEN` | API server listen address (default: `127.0.0.1:8080`) |
| `VGW_API_TOKEN` | Bearer token for API authentication (required for `--serve`) |

## Usage

### Interactive Mode
Run without arguments to launch the TUI:
```bash
vgw-manager
```

The Interactive Mode provides a rich terminal interface for all operations.

#### Navigation
*   **Arrow Keys / HJKL**: Navigate menus and lists.
*   **Enter**: Select item or confirm action.
*   **Esc**: Go back.
*   **Q / Ctrl+C**: Quit.

#### User Management
*   **List Users**: View all users.
    *   Press **c** to copy credentials to clipboard.
    *   Press **e** to edit a user.
    *   Press **d** to delete a user.
*   **Create User**: Setup new access/secret keys with specific roles (admin, user, userplus).

#### Bucket Management
*   **List Buckets**: View all buckets with real-time usage stats (Quota, Used, Available) and ownership status.
    *   Press **d** to delete a bucket.
    *   Press **p** (lowercase) to make a bucket **Public** (Read-only for everyone).
    *   Press **P** (uppercase) to make a bucket **Private** (Remove public policy).
*   **Create Bucket**: Create new ZFS-backed buckets with storage quotas.
*   **Change Owner**: Transfer bucket ownership to another user.

#### Operations
*   **Provision**: A wizard to create a User, a Bucket, and assign ownership/quotas in a single flow.
    *   Supports setting specific **UID**, **GID**, and **ProjectID** for advanced integration.
    *   Auto-generates **Secret Keys** if left blank.


#### Advanced Details
*   **Architecture**: `vgw-manager` operates on two layers:
    1.  **ZFS Layer**: Manages physical storage, datasets, and quotas directly on the host (requires root).
    2.  **VersityGW Layer**: Manages metadata, users, and ACLs via the Admin API.
*   **Roles**:
    *   `admin`: Full access to all operations.
    *   `user`: Standard S3 access to owned buckets.
    *   `userplus`: Can create buckets and manage own users.
*   **Public Buckets**: Setting a bucket to "Public" applies a policy granting `s3:GetObject` (Read-Only) to `*` (everyone) while maintaining full R/W access for the owner.

### CLI Commands

**User Management**
```bash
# Create User
vgw-manager --create-user --access "alice" --secret "securepass" --role "user" (optional --uid --gid)

# Delete User
vgw-manager --delete-user --access "alice"
```

**Bucket Management**
```bash
# Create Bucket with Quota
vgw-manager --create-bucket --bucket "archive" --quota "1T" --owner "alice"

# Make Bucket Public
vgw-manager --make-public --bucket "archive" --owner "alice"

# List Buckets (JSON output)
vgw-manager --list-buckets --json
```

**Provisioning**
```bash
# Provision User & Bucket
vgw-manager --provision --access "bob" --bucket "bob-data" --quota "500G"
```

### API Server

Run with `--serve` to start the HTTP API server:

```bash
vgw-manager --serve
```

**Configuration**: `apiToken` is required in the config file (or the `VGW_API_TOKEN` environment variable). The server refuses to start without it. `apiListen` defaults to `127.0.0.1:8080`.

#### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health check (no auth) |
| GET | `/v1/buckets` | List all buckets |
| POST | `/v1/buckets` | Create a bucket |
| DELETE | `/v1/buckets/{name}` | Delete a bucket |
| POST | `/v1/buckets/{name}/public` | Make bucket public |
| POST | `/v1/buckets/{name}/private` | Make bucket private |
| GET | `/v1/users` | List all users |
| GET | `/v1/users/{access}` | Get a single user |
| POST | `/v1/users` | Create a user |
| DELETE | `/v1/users/{access}` | Delete a user |
| POST | `/v1/provision` | Provision user + bucket + owner |

#### Examples

```bash
# Health check
curl http://127.0.0.1:8080/healthz

# List buckets
curl -H "Authorization: Bearer $VGW_API_TOKEN" http://127.0.0.1:8080/v1/buckets

# Create bucket
curl -X POST -H "Authorization: Bearer $VGW_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-bucket","quota":"1T","owner":"alice"}' \
  http://127.0.0.1:8080/v1/buckets

# Delete bucket
curl -X DELETE -H "Authorization: Bearer $VGW_API_TOKEN" \
  http://127.0.0.1:8080/v1/buckets/my-bucket

# Make bucket public
curl -X POST -H "Authorization: Bearer $VGW_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"owner":"alice"}' \
  http://127.0.0.1:8080/v1/buckets/my-bucket/public

# Make bucket private
curl -X POST -H "Authorization: Bearer $VGW_API_TOKEN" \
  http://127.0.0.1:8080/v1/buckets/my-bucket/private

# List users (secrets masked)
curl -H "Authorization: Bearer $VGW_API_TOKEN" http://127.0.0.1:8080/v1/users

# List users (show secrets)
curl -H "Authorization: Bearer $VGW_API_TOKEN" \
  "http://127.0.0.1:8080/v1/users?showSecrets=true"

# Get single user
curl -H "Authorization: Bearer $VGW_API_TOKEN" \
  http://127.0.0.1:8080/v1/users/alice

# Create user
curl -X POST -H "Authorization: Bearer $VGW_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"access":"alice","secret":"secret123","role":"user"}' \
  http://127.0.0.1:8080/v1/users

# Delete user
curl -X DELETE -H "Authorization: Bearer $VGW_API_TOKEN" \
  http://127.0.0.1:8080/v1/users/alice

# Provision user + bucket
curl -X POST -H "Authorization: Bearer $VGW_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"access":"bob","role":"user","bucket":"bob-data","quota":"500G"}' \
  http://127.0.0.1:8080/v1/provision
```

#### Example Responses

```json
// GET /v1/buckets
[
  {
    "name": "my-bucket",
    "mountpoint": "/tank/s3/buckets/my-bucket",
    "quota": "1T",
    "used": "100G",
    "available": "900G",
    "owner": "alice",
    "public": false
  }
]
```

```json
// POST /v1/provision
{
  "access": "bob",
  "secret": "auto-generated-secret",
  "role": "user",
  "userID": 0,
  "groupID": 0,
  "projectID": 0,
  "bucket": "bob-data",
  "quota": "500G",
  "owner": "bob",
  "secretGenerated": true
}
```

**Note**: Secrets are masked with `***` by default. Pass `?showSecrets=true` to reveal actual secret values.

## License

This project is licensed under the GNU General Public License v3.0 (GPLv3). See the [LICENSE](LICENSE) file for details.

[contributors-shield]: https://img.shields.io/github/contributors/monobilisim/vgw-manager.svg?style=for-the-badge
[contributors-url]: https://github.com/monobilisim/vgw-manager/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/monobilisim/vgw-manager.svg?style=for-the-badge
[forks-url]: https://github.com/monobilisim/vgw-manager/network/members
[stars-shield]: https://img.shields.io/github/stars/monobilisim/vgw-manager.svg?style=for-the-badge
[stars-url]: https://github.com/monobilisim/vgw-manager/stargazers
[issues-shield]: https://img.shields.io/github/issues/monobilisim/vgw-manager.svg?style=for-the-badge
[issues-url]: https://github.com/monobilisim/vgw-manager/issues
[license-shield]: https://img.shields.io/github/license/monobilisim/vgw-manager.svg?style=for-the-badge
[license-url]: https://github.com/monobilisim/vgw-manager/blob/master/LICENSE
[build-shield]: https://img.shields.io/github/actions/workflow/status/monobilisim/vgw-manager/build.yml?style=for-the-badge
[build-url]: https://github.com/monobilisim/vgw-manager/actions/workflows/build.yml
