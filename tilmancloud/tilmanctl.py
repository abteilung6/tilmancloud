from tilmancloud.cli.command import create_command, delete_command, tilmanctl
from tilmancloud.cli.resource import create_cluster, delete_cluster

create_command.add_command(create_cluster)
delete_command.add_command(delete_cluster)


if __name__ == '__main__':
    tilmanctl()
