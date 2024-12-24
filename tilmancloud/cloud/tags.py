
from datetime import datetime

from tilmancloud.shared.enums import StrEnum
from tilmancloud.shared.time import now


class CloudTag(StrEnum):
    create_time = 'tilmancloud-create-time'
    cluster_name = 'tilmancloud-cluster-name'


def build_aws_tags(cluster_name: str, create_time: datetime | None = None) -> dict:
    if create_time is None:
        create_time = now()
    return [
        {
            'Key': str(CloudTag.create_time),
            'Value': str(create_time.isoformat()),
        },
        {
            'Key': str(CloudTag.cluster_name),
            'Value': cluster_name,
        },
    ]


def build_resource_tags(name: str, cluster_name: str) -> dict:
    base_tags = build_aws_tags(cluster_name)
    vpc_tags = [
        {
            'Key': 'Name',
            'Value': name,
        },
    ]
    return base_tags + vpc_tags
