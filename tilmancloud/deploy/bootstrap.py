from tilmancloud.cloud.cloud_manager import AWSCloudConfig
from tilmancloud.cloud.subnet import CLUSTER_PUBLIC_SUBNET
from tilmancloud.cloud.vpc import CLUSTER_PUBLIC_VPC_CONFIG
from tilmancloud.deploy.cluster_action import ClusterAction
from tilmancloud.shared.logger import get_logger

LOG = get_logger(__name__)


class ClusterBootstraper(ClusterAction):
    """Responsible for bootstrapping the initial cluster infrastructure."""

    def __init__(self, cluster_name: str, cloud_config: AWSCloudConfig) -> None:
        super().__init__(cluster_name, cloud_config)

    def run(self) -> None:
        LOG.info('Bootstrapping cluster %s', self.cluster_name)
        self.create_cluster_vpcs()

    def create_cluster_vpcs(self) -> None:
        LOG.info('Boostrap cluster VPCs')
        vpc_id = self.cloud_manager.create_vpc(
            name=CLUSTER_PUBLIC_VPC_CONFIG.format_vpc_name(self.cluster_name),
            cidr_block=CLUSTER_PUBLIC_VPC_CONFIG.cidr_block,
            cluster_name=self.cluster_name,
        )
        self.cloud_manager.create_subnet(
            subnet=CLUSTER_PUBLIC_SUBNET,
            vpc_id=vpc_id,
            cluster_name=self.cluster_name,
        )
        LOG.info('Cluster VPCs bootstraped')
