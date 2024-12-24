import os

import click

from tilmancloud.cloud.cloud_manager import AWSCloudConfig, AWSCredentials
from tilmancloud.deploy.bootstrap import ClusterBootstraper
from tilmancloud.deploy.teardown import ClusterDestroyer
from tilmancloud.shared.enums import StrEnum


class ResourceType(StrEnum):
    cluster = 'cluster'


@click.argument('name', type=str)
@click.command(name=str(ResourceType.cluster))
def create_cluster(name: str) -> None:
    """Create a cluster."""
    aws_credentials = get_aws_credentials()
    cloud_config = AWSCloudConfig(credentials=aws_credentials, region='eu-west-1')
    bootstrapper = ClusterBootstraper(cluster_name=name, cloud_config=cloud_config)
    bootstrapper.run()


@click.argument('name', type=str)
@click.command(name=str(ResourceType.cluster))
def delete_cluster(name: str) -> None:
    """Delete a cluster."""
    aws_credentials = get_aws_credentials()
    cloud_config = AWSCloudConfig(credentials=aws_credentials, region='eu-west-1')
    destroyer = ClusterDestroyer(cluster_name=name, cloud_config=cloud_config)
    destroyer.run()


def get_aws_credentials() -> AWSCredentials:
    aws_access_key_id = os.environ['AWS_ACCESS_KEY_ID']
    aws_secret_access_key = os.environ['AWS_SECRET_ACCESS_KEY']
    if aws_access_key_id is None and not isinstance(aws_access_key_id, str):
        raise ValueError('aws_access_key_id is not set or is not a valid string')
    if aws_secret_access_key is None and not isinstance(aws_secret_access_key, str):
        raise ValueError('aws_secret_access_key is not set or is not a valid string')
    return AWSCredentials(
        aws_key_id=aws_access_key_id,
        aws_secret_access_key=aws_secret_access_key,
    )
