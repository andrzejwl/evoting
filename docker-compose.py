"""
generate docker-compose dynamically
"""

from typing import List
import yaml

yaml.Dumper.ignore_aliases = lambda *args : True # for some reason pyyaml is dumping pointer references instead of arrays without this line

class Compose(object):
    def __init__(self, version='3.9') -> None:
        self.root = {
            'version': version,
            'services': {}
        }
        super().__init__()

    def add_service(self, name: str, image: str = None, ports: List[str] = None, env_vars: List['str'] = None, 
                    depends_on: List['str'] = None, entrypoint: str = None, container_name: str = None):
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

        if container_name:
            self.root['services'][name]['container_name'] = container_name

    def dump(self, filename='docker-compose-py.yml'):
        f = open(filename, 'w+')
        yaml.dump(self.root, f)
        f.close()
        
        # remove single quotes from entrypoint
        f = open(filename, 'r+')
        contents = f.read()
        contents = contents.replace("'[", "[")
        contents = contents.replace("]'", "]")
        f.seek(0)
        f.write(contents)
        f.truncate()
        f.close()
    

compose = Compose()
# compose.add_service(name='node-discovery', image='evoting_node-discovery', ports=['9999:9999',])
# compose.add_service(name='connector', image='evoting_connector', ports=['1234:1234',], env_vars=['ND_ADDR=node-discovery:9999'])

dependencies = ['node-discovery']

# blockchain nodes
# for i in range(10):
#     deps = dependencies[:]
#     if i != 0:
#         deps.append(f'node-{i}')

#     compose.add_service(
#         name=f'node-{i+1}', 
#         image='evoting_node', 
#         ports=[f'{1337+i}:{1337+i}',], 
#         depends_on=deps, 
#         env_vars=[
#             f'HOSTNAME=node-{i+1}', 
#             'DISCOVERY_ADDR=node-discovery:9999',
#         ],
#         entrypoint=f'["/evoting", "-consensus=pbft", "-port={1337+i}"]',
#     )

# # client nodes
# for i in range(3):
#     compose.add_service(
#         name=f'client-{i+1}', 
#         image='evoting_node', 
#         ports=[f'{2001+i}:{2001+i}',], 
#         depends_on=['node-discovery'], 
#         env_vars=[
#             f'HOSTNAME=client-{i+1}', 
#             'DISCOVERY_ADDR=node-discovery:9999',
#         ],
#         entrypoint=f'["/evoting", "-client_mode=true", "-port={2001+i}"]',
#     )

# PoW

# root node
compose.add_service(
    name='node-1',
    image='evoting_node',
    ports=['1337:1337',],
    entrypoint='["/evoting", "-consensus=pow", "-root=true"]',
    env_vars=[
        'DOCKER=1',
        'HOSTNAME=node-1',
        'PORT=1337',
    ],
    container_name='pow_node-1'
)

for i in range(1,10):
    compose.add_service(
        name=f'node-{i+1}', 
        image='evoting_node', 
        ports=[f'{1337+i}:{1337+i}',], 
        depends_on=[f'node-{i}'], 
        env_vars=[
            'DOCKER=1',
            f'HOSTNAME=node-{i+1}',
            f'PORT={1337+i}',
            'PEER_HOSTNAME=node-1',
            'PEER_PORT=1337',
        ],
        entrypoint=f'["/evoting", "-consensus=pow"]',
        container_name=f'pow_node-{i+1}'
    )

compose.dump('compose-pow-py.yml')    
