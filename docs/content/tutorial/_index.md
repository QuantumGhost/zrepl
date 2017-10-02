---
title: "Tutorial"
weight: 1
---

This tutorial shows how zrepl can be used to implement a ZFS-based pull backup.

We assume the following scenario:

* Production server `app-srv` with filesystems to back up:
    * `zroot/var/db`
    * `zroot/usr/home` and all its child filesystems
    * **except** `zroot/usr/home/paranoid` belonging to a user doing backups themselves
* Backup server `backup-srv` with
    * Filesystem `storage/zrepl/pull/app-srv` + children dedicated to backups of `app-srv`

Our backup solution should fulfill the following requirements:

* Periodically snapshot the filesystems on `app-srv` *every 10 minutes*
* Incrementally replicate these snapshots to `storage/zrepl/pull/app-srv/*` on `backup-srv`
* Keep only very few snapshots on `app-srv` to save disk space
* Keep a fading history (24 hourly, 30 daily, 6 monthly) of snapshots on `backup-srv`

## Analysis

We can model this situation as two jobs:

* A **source job** on `app-srv`
    * Creates the snapshots
    * Keeps a short history of snapshots to enable incremental replication to `backup-srv`
    * Accepts connections from `backup-srv`
* A **pull job** on `backup-srv`
    * Connects to the `zrepl daemon` process on `app-srv`
    * Pulls the snapshots to `storage/zrepl/pull/app-srv/*`
    * Fades out snapshots in `storage/zrepl/pull/app-srv/*` as they age

{{%expand "Side note: why doesn't the **pull job** create the snapshots?" %}}

As is the case with all distributed systems, the link between `app-srv` and `backup-srv` might be down for an hour or two.
We do not want to sacrifice our required backup resolution of 10 minute intervals for a temporary connection outage.

When the link comes up again, `backup-srv` will happily catch up the 12 snapshots taken by `app-srv` in the meantime, without
a gap in our backup history.
{{%/expand%}}

## Install zrepl

Follow the [OS-specific installation instructions]({{< relref "install/_index.md" >}}) and come back here.

## Configure `backup-srv`

We define a **pull job** named `pull_app-srv` in the [main configuration file]({{< relref "install/_index.md#configuration-files" >}} ):

```yaml
jobs:
- name: pull_app-srv
  type: pull
  connect:
    type: ssh+stdinserver
    host: app-srv.example.com
    user: root
    port: 22
    identity_file: /etc/zrepl/ssh/app-srv
  interval: 10m
  mapping: {
    "<":"storage/zrepl/pull/app-srv"
  }
  initial_repl_policy: most_recent
  snapshot_prefix: zrepl_pull_backup_
  prune:
    policy: grid
    grid: 1x1h(keep=all) | 24x1h | 35x1d | 6x30d
```

The `connect` section instructs zrepl to use the `stdinserver` transport: instead of directly exposing zrepl on `app-srv`
to the internet, `backup-srv` starts the `zrepl stdinserver` subcommand on `app-srv` via SSH.
(You can learn more about what happens [here]({{< relref "configuration/transports.md#stdinserver" >}}), or just continue following this tutorial.)

Thus, we need to create the SSH key pair `/etc/zrepl/ssh/app-srv{,.pub}` which identifies `backup-srv` toward `app-srv`.
Execute the following commands on `backup-srv` as the root user:

```bash
cd /etc/zrepl
mkdir -p ssh
chmod 0700 ssh
ssh-keygen -t ed25519 -N '' -f /etc/zrepl/ssh/app-srv
```
You can learn more about the [**pull job** format here]({{< relref "configuration/jobs.md#pull" >}}) but for now we are good to go.

## Configure `app-srv`

We define a corresponding **source job** named `pull_backup` in the [main configuration file]({{< relref "install/_index.md#configuration-files" >}})
`zrepl.yml`:

```yaml
jobs:

- name: pull_backup
  type: source
  serve:
    type: stdinserver
    client_identity: backup-srv.example.com
  datasets: {
    "zroot/var/db": "ok",
    "zroot/usr/home<": "ok",
    "zroot/usr/home/paranoid": "!",
  }
  snapshot_prefix: zrepl_pull_backup_
  interval: 10m
  prune:
    policy: grid
    grid: 1x1d(keep=all)

```

The `serve` section corresponds to the `connect` section in the configuration of `backup-srv`.

As mentioned before, the SSH key `app-srv.pub` created in the section before identifies `backup-srv` toward `app-srv`.
We enforce that by limiting `backup-srv` to execute exactly `zrepl stdinserver CLIENT_IDENTITY` when connecting to `app-srv`.

Open `/root/.ssh/authorized_keys` and add either of the the following lines.<br />

```
# for OpenSSH >= 7.2
command="zrepl stdinserver backup-srv.example.com",restrict PULLING_SSH_KEY
# for older OpenSSH versions
command="zrepl stdinserver backup-srv.example.com",no-port-forwarding,no-X11-forwarding,no-pty,no-agent-forwarding,no-user-rc  PULLING_SSH_KEY
```

{{% notice info %}}
Replace PULLING_SSH_KEY with the contents of `app-srv.pub`.<br/>
The entries **must** be on a single line, including the replaced PULLING_SSH_KEY.
{{% /notice %}}

Again, you can learn more about the [**source job** format here]({{< relref "configuration/jobs.md#source" >}}).

## Apply Configuration Changes

We need to restart the zrepl daemon on **both** `app-srv` and `backup-srv`.

This is [OS-specific]({{< relref "install/_index.md#restarting" >}}).

## Watch it Work

A common setup is to `watch` the log output and `zfs list` of snapshots on both machines.

If you like tmux, here is a handy script that works on FreeBSD:

```bash
pkg install gnu-watch tmux
tmux new-window
tmux split-window "tail -f /var/log/zrepl.log"
tmux split-window "gnu-watch 'zfs list -t snapshot -o name,creation -s creation | grep zrepl_pull_backup_'"
tmux select-layout tiled
```

The Linux equivalent might look like this

```bash
# make sure tmux is installed & let's assume you use systemd + journald
tmux new-window
tmux split-window "journalctl -f -u zrepl.service"
tmux split-window "watch 'zfs list -t snapshot -o name,creation -s creation | grep zrepl_pull_backup_'"
tmux select-layout tiled
```

## Summary

Congratulations, you have a working pull backup. Where to go next?

* Read more about [configuration format, options & job types]({{< relref "configuration/_index.md" >}})
* Learn about [implementation details]({{<relref "impl/_index.md" >}}) of zrepl.



