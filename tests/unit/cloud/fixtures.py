import pytest
from boto3 import Session
from botocore.stub import Stubber

from tilmancloud.cloud.cloud_manager import AWSCloudConfig, AWSCloudManager, AWSCredentials


class AWSCloudManagerMock(AWSCloudManager):
    def __init__(self, config: AWSCloudConfig) -> None:
        super().__init__(config)
        self.stubbers: dict[str, Stubber] = {}

    def _create_botocore_session(self) -> None:
        if self._botocore_session is None:
            self._botocore_session = Session()

    def _create_botocore_client(self, service_name: str):  # noqa: ANN202
        if service_name not in self.stubbers:
            stubber = Stubber(super()._create_botocore_client(service_name=service_name))
            self.stubbers[service_name] = stubber
            stubber.activate()
        return self.stubbers[service_name].client

    def get_stubber(self, service_name: str) -> Stubber:
        if service_name not in self.stubbers:
            self._create_botocore_client(service_name=service_name)
        return self.stubbers[service_name]


@pytest.fixture(name="aws_cloud_config")
def fixture_aws_cloud_config() -> AWSCloudConfig:
    return AWSCloudConfig(
        credentials=AWSCredentials(
            aws_key_id='ident',
            aws_secret_access_key='secret',
        ),
        region='eu-west-1',
    )


@pytest.fixture(name="aws_cloud_manager")
def fixture_aws_cloud_manager(aws_cloud_config: AWSCloudConfig) -> AWSCloudManagerMock:
    return AWSCloudManagerMock(config=aws_cloud_config)
