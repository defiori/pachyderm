import subprocess
import time
import platform
import json

from tornado.httpclient import AsyncHTTPClient, HTTPClientError
from tornado import locks

from .pachyderm import MountInterface
from .log import get_logger
from .env import SIDECAR_MODE

lock = locks.Lock()
MOUNT_SERVER_PORT = 9002


class MountServerClient(MountInterface):
    """Client interface for the mount-server backend."""

    def __init__(
        self,
        mount_dir: str,
    ):
        self.client = AsyncHTTPClient()
        self.mount_dir = mount_dir
        self.address = f"http://localhost:{MOUNT_SERVER_PORT}"

    async def _is_mount_server_running(self):
        get_logger().debug("Checking if mount server running...")
        try:
            await self.health()
        except Exception as e:
            get_logger().debug(f"Unable to hit server at {self.address}")
            get_logger().debug(e)
            return False
        get_logger().debug(f"Able to hit server at {self.address}")
        return True

    def _unmount(self):
        if platform.system() == "Linux":
            subprocess.run(["bash", "-c", f"fusermount -uzq {self.mount_dir}"])
        else:
            subprocess.run(
                ["bash", "-c", f"umount {self.mount_dir}"],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
            )

    async def _ensure_mount_server(self):
        """
        When we first start up, we might not have auth configured. So just try
        re-launching the mount-server on every command! If the port is already
        bound, it will exit straight away. If it's not, it might start up
        successfully with the updated config.
        """
        # TODO: add --socket and --log-file stdout args
        # TODO: add better error handling
        if await self._is_mount_server_running():
            return True

        if SIDECAR_MODE:
            get_logger().debug("Kubernetes is responsible for running mount server")
            return False

        get_logger().info("Starting mount server...")
        async with lock:
            if not await self._is_mount_server_running():
                self._unmount()
                subprocess.Popen(
                    [
                        "bash",
                        "-c",
                        "set -o pipefail; "
                        + f"mount-server --mount-dir {self.mount_dir}"
                        + " >> /tmp/mount-server.log 2>&1",
                    ]
                )

                tries = 0
                get_logger().debug("Waiting for mount server...")
                while not await self._is_mount_server_running():
                    time.sleep(1)
                    tries += 1

                    if tries == 10:
                        get_logger().debug("Unable to start mount server...")
                        return False

        return True


    async def list_repos(self):
        await self._ensure_mount_server()
        response = await self.client.fetch(f"{self.address}/repos")
        return response.body

    async def list_mounts(self):
        await self._ensure_mount_server()
        response = await self.client.fetch(f"{self.address}/mounts")
        return response.body
        
    async def mount(self, body):
        await self._ensure_mount_server()
        response = await self.client.fetch(
            f"{self.address}/_mount",
            method="PUT",
            body=json.dumps(body),
        )
        return response.body

    async def unmount(self, body):
        await self._ensure_mount_server()
        response = await self.client.fetch(
            f"{self.address}/_unmount",
            method="PUT",
            body=json.dumps(body),
        )
        return response.body

    async def commit(self, body):
        await self._ensure_mount_server()
        pass

    async def unmount_all(self):
        await self._ensure_mount_server()
        response = await self.client.fetch(
            f"{self.address}/_unmount_all",
            method="PUT",
            body="{}"
        )
        return response.body

    async def mount_datums(self, body):
        await self._ensure_mount_server()
        response = await self.client.fetch(
            f"{self.address}/_mount_datums",
            method="PUT",
            body=json.dumps(body),
        )
        return response.body

    async def show_datum(self, slug):
        await self._ensure_mount_server()
        slug = '&'.join(f"{k}={v}" for k,v in slug.items() if v is not None)
        response = await self.client.fetch(
            f"{self.address}/_show_datum?{slug}",
            method="PUT",
            body="{}",
        )
        return response.body

    async def get_datums(self):
        await self._ensure_mount_server()
        response = await self.client.fetch(f"{self.address}/datums")
        return response.body

    async def config(self, body=None):
        await self._ensure_mount_server()
        if body is None:
            try:
                response = await self.client.fetch(f"{self.address}/config")
            except HTTPClientError as e:
                if e.code == 404:
                    return json.dumps({"cluster_status": "INVALID"})
                raise e
        else:
            response = await self.client.fetch(
                f"{self.address}/config", method="PUT", body=json.dumps(body)
            )

        return response.body

    async def auth_login(self):
        await self._ensure_mount_server()
        response = await self.client.fetch(
            f"{self.address}/auth/_login", method="PUT", body="{}"
        )
        return response.body

    async def auth_logout(self):
        await self._ensure_mount_server()
        return await self.client.fetch(
            f"{self.address}/auth/_logout", method="PUT", body="{}"
        )

    async def health(self):
        response = await self.client.fetch(f"{self.address}/health")
        return response.body
