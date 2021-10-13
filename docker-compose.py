"""
generate docker-compose dynamically
"""

from typing import Dict, List
import yaml

class Compose(object):
    def __init__(self, version='3.9') -> None:
        self.root = {
            'version': version,
            'services': {}
        }
        super().__init__()

    def add_service(self, name: str, image: str = None, ports: List[str] = None, env_vars: List['str'] = None, depends_on: List['str'] = None, entrypoint: List = None):
        self.root['services'][name] = {}

        if image:
            self.root['services'][name]['image'] = image

        if ports:
            self.root['services'][name]['ports'] = ports

        if env_vars:
            self.root['services'][name]['environment'] = env_vars

        if depends_on:
            self.root['services'][name]['depends_on'] = depends_on

        if entrypoint:
            self.root['services'][name]['entrypoint'] = entrypoint

    def dump(self, filename='docker-compose-py.yml'):
        f = open(filename, 'w+')
        yaml.dump(self.root, f)
        f.close()


compose = Compose()
compose.add_service(name='node-discovery', image='evoting_node-discovery', ports=['9999:9999',])
compose.add_service(name='connector', image='evoting_connector', ports=['1234:1234',], env_vars=['ND_ADDR=node-discovery:9999'])
for i in range(10):
    compose.add_service(
        name=f'node-{i+1}', 
        image='evoting_node', 
        ports=[f'{1337+i}:{1337+i}',], 
        depends_on=['node-discovery'], 
        env_vars=[f'HOSTNAME=node-{i+1}', 
        'DISCOVERY_ADDR=node-discovery:9999'],
        entrypoint=f'["/evoting", "-consensus=pbft", "-port={1337+i}"]',
    )
compose.dump()    
