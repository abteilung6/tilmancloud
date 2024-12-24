from dataclasses import dataclass
from ipaddress import IPv4Network


@dataclass(frozen=True)
class Subnet:
    scope: str
    availability_zone: str
    cidr_block: IPv4Network

    def format_subnet_name(self, cluster_name: str) -> str:
        return f'{cluster_name}-{self.scope}-subnet'



CLUSTER_PUBLIC_SUBNET = Subnet(scope="public", availability_zone="eu-west-1a", cidr_block=IPv4Network("10.1.0.0/20"))
