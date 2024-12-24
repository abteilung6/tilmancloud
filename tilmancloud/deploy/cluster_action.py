from abc import ABC, abstractmethod
from functools import cached_property

from tilmancloud.cloud.cloud_manager import AWSCloudConfig, AWSCloudManager


class ClusterAction(ABC):

    def __init__(self, cluster_name: str, cloud_config: AWSCloudConfig) -> None:
        self.cluster_name = cluster_name
        self.cloud_config = cloud_config

    @abstractmethod
    def run(self) -> None:
        """Perform the cluster action."""
        pass

    @cached_property
    def cloud_manager(self) -> AWSCloudManager:
        return AWSCloudManager(self.cloud_config)
