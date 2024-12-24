import click

from tilmancloud.shared.enums import StrEnum


class CommandGroup(StrEnum):
    create = 'create'
    delete = 'delete'


@click.group(name='tilmanctl')
def tilmanctl() -> None:
    ...

@tilmanctl.group(name=str(CommandGroup.create))
def create_command() -> None:
    """Create a resource."""
    ...

@tilmanctl.group(name=str(CommandGroup.delete))
def delete_command() -> None:
    """Delete a resource."""
    ...
