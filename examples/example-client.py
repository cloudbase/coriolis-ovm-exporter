#!/usr/bin/env python3

import os
import sys
import requests
import urllib.parse as urlparse

from prettytable import PrettyTable

import urllib3
urllib3.disable_warnings()

class Client(object):

    def __init__(self, endpoint, token):
        self._token = token
        self._endpoint = endpoint
        self._cli_obj = None

    @property
    def _cli(self):
        if self._cli_obj is not None:
            return self._cli_obj
        sess = requests.Session()
        sess.headers = self._get_headers()
        sess.verify = False
        return sess

    @classmethod
    def login(cls, endpoint, username, password):
        loginURL = urlparse.urljoin(endpoint, "/api/v1/auth/login") 
        ret = requests.post(
                loginURL, verify=False,
                json={"username": username, "password": password})
        ret.raise_for_status()
        token = ret.json().get("token", None)
        if token is None:
            raise ValueError("could not get access token")
        return cls(endpoint, token=token)

    def _get_headers(self):
        headers = {
            "Authorization": "Bearer %s" % self._token
        }
        return headers

    def get_vms(self):
        url = urlparse.urljoin(self._endpoint, "/api/v1/vms/")

        ret = self._cli.get(
                url)
        ret.raise_for_status()
        return ret.json()

    def get_vm(self, vmID):
        url = urlparse.urljoin(self._endpoint, "/api/v1/vms/%s" % vmID)

        ret = self._cli.get(
                url)
        ret.raise_for_status()
        return ret.json()

    def create_snapshot(self, vmID):
        url = urlparse.urljoin(
                self._endpoint,
                "/api/v1/vms/%s/snapshots/" % vmID)
        ret = self._cli.post(
                url)
        ret.raise_for_status()
        return ret.json()

    def get_snapshots(self, vmID):
        url = urlparse.urljoin(
                self._endpoint,
                "/api/v1/vms/%s/snapshots/" % vmID)
        ret = self._cli.get(
                url)
        ret.raise_for_status()
        return ret.json()

    def get_snapshot(self, vmID, snapshotID, compare_to=None, squash=False):
        url = urlparse.urljoin(
                self._endpoint,
                "/api/v1/vms/%s/snapshots/%s/" % (vmID, snapshotID))

        params = {
            "squashChunks": squash
        }
        if compare_to is not None:
            params["compareTo"] = compare_to
        ret = self._cli.get(
                url, params=params)
        ret.raise_for_status()
        return ret.json()

    def get_disk_size(self, vmID, snapID, diskID):
        url = urlparse.urljoin(
                self._endpoint,
                "/api/v1/vms/%s/snapshots/%s/disks/%s" % (
                    vmID, snapID, diskID))
        ret = self._cli.head(
                url)
        ret.raise_for_status()
        length = ret.headers.get("content-length", None)
        if length is None:
            raise Exception("failed to get content length")
        return int(length)

    def download_chunk(self, vmID, snapID, diskID, offset, length):
        url = urlparse.urljoin(
                self._endpoint,
                "/api/v1/vms/%s/snapshots/%s/disks/%s" % (
                    vmID, snapID, diskID))
        start = offset
        end = offset + length - 1

        headers = self._get_headers()
        headers["Range"] = "bytes=%s-%s" % (start, end)
        ret = requests.get(url, headers=headers, verify=False)
        ret.raise_for_status()
        return ret.content


def _get_creds_from_env():
    endpoint = os.environ.get("OVM_EXPORTER_ENDPOINT", None)
    username = os.environ.get("OVM_EXPORTER_USERNAME", None)
    password = os.environ.get("OVM_EXPORTER_PASSWORD", None)

    if not all([endpoint, username, password]):
        print(
            "OVM_EXPORTER_ENDPOINT, OVM_EXPORTER_USERNAME and"
            " OVM_EXPORTER_PASSWORD environment variables must"
            " be set")
        sys.exit(1)
    return {
        "endpoint": endpoint,
        "username": username,
        "password": password
    }

def _get_client():
    creds = _get_creds_from_env()
    cli = Client.login(
        endpoint=creds["endpoint"],
        username=creds["username"],
        password=creds["password"])
    return cli


def get_vms():
    cli = _get_client()
    vms = cli.get_vms()
    x = PrettyTable()
    x.field_names = ["ID", "Friendly Name", "Snapshots"]
    for vm in vms:
        x.add_row(
            [
                vm["name"], vm["friendly_name"],
                "\n".join(vm["snapshots"])])
    x.align = "l"
    print(x)


def get_snapshots(vmID):
    cli = _get_client()
    snapshots = cli.get_snapshots(vmID)
    x = PrettyTable()
    x.field_names = ["Snapshot ID", "VM ID", "Disks"]
    for snap in snapshots:
        x.add_row(
            [
                snap["id"], snap["vm_id"],
                "\n".join([disk["name"] for disk in snap["disks"]])])
    x.align = "l"
    print(x)


def get_snapshot(vmID, snapID):
    cli = _get_client()
    snap = cli.get_snapshot(vmID, snapID, squash=True)

    x = PrettyTable()
    x.field_names = ["Snapshot ID", "VM ID", "Disks"]
    x.add_row(
        [
            snap["id"], snap["vm_id"],
            "\n".join([disk["name"] for disk in snap["disks"]])])
    x.align = "l"
    print(x)


def download_vm_disks(vmID, snapID, out_dir, diffs_from=None):
    """Downloads all VM disks belonging to snapID to out_dir.
    The diffs_from option allows you to specify a previous snapshot ID. Only
    changes from this snapshot ID will be downloaded and written to appropriate
    offsets.
    """
    cli = _get_client()
    kw = {
        "squash": False,
    }

    if diffs_from is not None:
        kw["compare_to"] = diffs_from
    snapshot = cli.get_snapshot(vmID, snapID, **kw)

    if os.path.isdir(out_dir) is False:
        os.makedirs(out_dir)

    for disk in snapshot["disks"]:
        size = cli.get_disk_size(vmID, snapID, disk["name"])
        download_path = os.path.join(out_dir, disk["name"].lstrip("/"))
        if os.path.exists(download_path) is False:
            fd = open(download_path, 'wb')
            fd.truncate(size)
            fd.seek(0)
        else:
            fd = open(download_path, 'r+b')

        for chunk in disk["chunks"]:
            fd.seek(chunk["start"])
            data = cli.download_chunk(
                vmID, snapID, disk["name"], chunk["start"], chunk["length"])
            fd.write(data)
        fd.close()

