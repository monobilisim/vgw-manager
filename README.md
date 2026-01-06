[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![GPL License][license-shield]][license-url]
[![Build Status][build-shield]][build-url]


<div align="center">
  <a href="https://mono.tr/">
    <img src="https://monobilisim.com.tr/images/mono-bilisim.svg" width="340"/>
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
