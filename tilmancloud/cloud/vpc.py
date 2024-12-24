
from dataclasses import dataclass
from ipaddress import IPv4Network


@dataclass(frozen=True)
class VpcConfig:
    scope: str
    cidr_block: IPv4Network

    def format_vpc_name(self, cluster_name: str) -> str:
        return f'{cluster_name}-{self.scope}-vpc'


CLUSTER_PUBLIC_VPC_CONFIG = VpcConfig(scope="public", cidr_block=IPv4Network("10.1.0.0/16"))
