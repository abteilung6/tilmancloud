from tilmancloud.cloud.cloud_manager import AWSCloudConfig
from tilmancloud.deploy.cluster_action import ClusterAction
from tilmancloud.shared.logger import get_logger

LOG = get_logger(__name__)


class ClusterDestroyer(ClusterAction):
    """Responsible for destroying the cluster infrastructure."""

    def __init__(self, cluster_name: str, cloud_config: AWSCloudConfig) -> None:
        super().__init__(cluster_name, cloud_config)

    def run(self) -> None:
        LOG.info('Destroying cluster %s', self.cluster_name)
        self.delete_cluster_vpcs()
        LOG.info('Destroyed cluster %s', self.cluster_name)

    def delete_cluster_vpcs(self) -> None:
        LOG.info('Deleting VPCs for %s', self.cluster_name)
        vpc_ids = self.cloud_manager.find_vpcs(cluster_name=self.cluster_name)
        LOG.info('Found %i VPCs for cluster %s', len(vpc_ids), self.cluster_name)
        for vpc_id in vpc_ids:
            self.cloud_manager.delete_vpc(vpc_id)
        LOG.info('Deleted VPCs for %s', self.cluster_name)
