

from ipaddress import IPv4Network

from tests.unit.cloud.fixtures import (  # noqa: F401
    AWSCloudManagerMock,
    fixture_aws_cloud_config,
    fixture_aws_cloud_manager,
)


def test_create_vpc(aws_cloud_manager: AWSCloudManagerMock) -> None:
    create_vpc_response = {
        "Vpc": {
            "VpcId": "vpc_id",
        },
    }
    stubber = aws_cloud_manager.get_stubber('ec2')
    stubber.add_response(method="create_vpc", service_response=create_vpc_response)
    response = aws_cloud_manager.create_vpc(
        name='vpc_name',
        cidr_block=IPv4Network("192.168.1.0/24"),
        cluster_name='cluster_name',
    )
    assert response == "vpc_id"
