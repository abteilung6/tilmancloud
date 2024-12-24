from dataclasses import dataclass
from functools import cached_property
from ipaddress import IPv4Network

from boto3 import Session
from mypy_boto3_ec2.client import EC2Client

from tilmancloud.cloud.subnet import Subnet
from tilmancloud.cloud.tags import CloudTag, build_resource_tags
from tilmancloud.shared.logger import get_logger
from tilmancloud.shared.types import VpcId

LOG = get_logger(__name__)


@dataclass(frozen=True)
class AWSCredentials:
    aws_key_id: str
    aws_secret_access_key: str


@dataclass(frozen=True)
class AWSCloudConfig:
    credentials: AWSCredentials
    region: str


class AWSCloudManager:
    def __init__(self, config: AWSCloudConfig) -> None:
        self.config = config
        self._botocore_session: Session | None = None

    def _create_botocore_session(self) -> None:
        if self._botocore_session is None:
            self._botocore_session = Session(
                aws_access_key_id=self.config.credentials.aws_key_id,
                aws_secret_access_key=self.config.credentials.aws_secret_access_key,
            )

    def _create_botocore_client(self, service_name: str) -> Session:
        self._create_botocore_session()
        assert self._botocore_session is not None
        return self._botocore_session.client(service_name=service_name, region_name=self.config.region)


    @cached_property
    def ec2_client(self) -> EC2Client:
        return self._create_botocore_client('ec2')


    def create_vpc(self, name: str, cidr_block: IPv4Network, cluster_name: str) -> VpcId:
        LOG.info('Creating VPC %s with cidr_block=%s', name, cidr_block)
        vpc_tags = build_resource_tags(name, cluster_name)
        create_vpc_result = self.ec2_client.create_vpc(
            CidrBlock=str(cidr_block),
            TagSpecifications=[
                {
                    'ResourceType': 'vpc',
                    'Tags': vpc_tags,
                },
            ],
        )
        vpc_id = create_vpc_result['Vpc']['VpcId']
        LOG.info('Created VPC %s with vpc_id=%s', name, vpc_id)
        return VpcId(vpc_id)


    def create_subnet(self, subnet: Subnet, vpc_id: VpcId, cluster_name: str) -> None:
        subnet_name = subnet.format_subnet_name(cluster_name)
        LOG.info('Creating subnet %s with vpc_id=%s', subnet_name, vpc_id)
        subnet_tags = build_resource_tags(subnet_name, cluster_name)
        self.ec2_client.create_subnet(
            VpcId=vpc_id,
            CidrBlock=str(subnet.cidr_block),
            AvailabilityZone=subnet.availability_zone,
            TagSpecifications=[
                {
                    'ResourceType': 'subnet',
                    'Tags': subnet_tags,
                },
            ],
        )
        LOG.info('Created subnet %s', subnet_name)


    def delete_vpc(self, vpc_id: VpcId) -> None:
        LOG.info('Deleting VPC with vpc_id=%s', vpc_id)
        self.ec2_client.delete_vpc(VpcId=vpc_id)
        LOG.info('Deleted VPC with vpc_id=%s', vpc_id)


    def find_vpcs(self, cluster_name: str) -> list[VpcId]:
        describe_vpcs_result = self.ec2_client.describe_vpcs(
            Filters=[
                {
                    'Name': f"tag:{str(CloudTag.cluster_name)}",
                    'Values': [cluster_name],
                },
            ],
        )
        return [VpcId(entry["VpcId"]) for entry in describe_vpcs_result["Vpcs"]]
