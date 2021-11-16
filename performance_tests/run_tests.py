import argparse
import docker
import requests
import subprocess
import time
import xlsxwriter
from typing import Dict, List
from pathlib import Path
from datetime import datetime

from xlsxwriter.workbook import Workbook


def get_containers(): 
    client = docker.from_env()
    return client.containers.list()


def restart_containers(compose_path: str, compose_file: str):
    # Docker SDK doesn't support docker-compose so have to do it manually 
    process = subprocess.run(['docker-compose', '-f', compose_file, 'restart'], cwd=compose_path, stdout=subprocess.DEVNULL)


def start_compose(compose_path: str, compose_file: str):
    process = subprocess.run(['docker-compose', '-f', compose_file, 'up', '-d'], cwd=compose_path, stdout=subprocess.DEVNULL)


def close_compose(compose_path: str, compose_file: str):
    process = subprocess.run(['docker-compose', '-f', compose_file, 'stop'], cwd=compose_path, stdout=subprocess.DEVNULL)


def body(consensus: str, i: int, toId = None)->Dict:
    if consensus == 'pow':
        body = {
            'Token': f'token {i}',
            'ToId': f'transaction {i}' if not toId else toId,
        }
        return body
    elif consensus == 'pbft':
        body = {
                'transactions': [
                    {
                        'Token': f'token {i}',
                        'ToId': f'transaction {i}' if not toId else toId,
                    }
                ]
            }
        return body
    else:
        raise Exception('incorrect consensus protocol')


def test_performance(transactions: int, node_address: str, consensus: str)->float:
    """
    Returns time taken to submit all of the transactions.
    """
    parties = ["voting party 1", "voting party 2"]
    start = time.time()
    url = f'{node_address}/transaction/create' if consensus == 'pow' else f'{node_address}/new-request'
    
    for token in parties:
        for i in range(transactions//len(parties)):
            r = requests.post(url, json=body(consensus, i, token))
    end = time.time()

    return end-start


def container_statistics():
    client = docker.from_env()
    for i in client.containers.list():
        # print(i.stats(stream=False))
        print(cpu_usage_percent(i.stats(stream=False)))


def cpu_usage_percent(stats):
    usage_delta = stats['cpu_stats']['cpu_usage']['total_usage'] - stats['precpu_stats']['cpu_usage']['total_usage']
    sys_delta = stats['cpu_stats']['system_cpu_usage'] - stats['precpu_stats']['system_cpu_usage']
    len_cpu = len(stats['cpu_stats']['cpu_usage']['percpu_usage'])
    percentage = (usage_delta / sys_delta) * len_cpu * 100
    return round(percentage, 2)


def query_usage(query: str, name: str, start: int, end: int):
    response = requests.get(f"http://localhost:9090/api/v1/query_range?query={query}{{name='{name}'}}&start={start}&end={end}&step=1")
    if response.status_code != 200:
        raise Exception(f'Failed to to fetch query data for {name}')
    server_response = response.json()
    try:
        return server_response['data']['result'][0]['values']
    except IndexError:
        return server_response['data']['result']

def get_memory_usage(name: str, start: int, end: int):
    response = requests.get(f"http://localhost:9090/api/v1/query_range?query=container_memory_usage_bytes{{name='{name}'}}&start={start}&end={end}&step=1")
    if response.status_code != 200:
        raise Exception(f'Failed to to fetch memory data for {name}')
    server_response = response.json()
    try:
        return server_response['data']['result'][0]['values']
    except IndexError:
        return server_response['data']['result']

def get_10s_cpu_usage(name: str, start: int, end: int):
    response = requests.get(f"http://localhost:9090/api/v1/query_range?query=container_cpu_load_average_10s{{name='{name}'}}&start={start}&end={end}&step=1")
    if response.status_code != 200:
        print(response.json())
        raise Exception(f'Failed to to fetch CPU data for {name}')
    server_response = response.json()
    try:
        return server_response['data']['result'][0]['values']
    except IndexError:
        return server_response['data']['result']


def dump_data_to_xlsx(notebooks: Dict[str, xlsxwriter.Workbook], data: Dict):
    nodes = list(data.keys())
    for label, notebook in notebooks.items():
        worksheet = notebook.add_worksheet()

        for col in range(len(nodes)):
            node = nodes[col]
            worksheet.write(0, col*3, node)
            
            row = 1
            for ts, val in data[node][label]:
                worksheet.write(row, col*3, ts)
                worksheet.write(row, col*3+1, val)
                row += 1


def dump_round_data(notebook: xlsxwriter.Workbook, data: List):
    params = list(data[0].keys())
    row = 0
    worksheet = notebook.add_worksheet()

    for round in data:
        for label, val in round.items():
            worksheet.write(row, 0, label)
            worksheet.write(row, 1, val)
            row += 1
        row += 1


def start_tests_for_consensus(consensus: str, transactions: int, rounds: int, node_address: str, number_of_nodes: int):
    """
    Runs tests in multiple rounds and dumps data to xlsx files.
    """
    
    compose_file = 'compose-pow.yml' if consensus == 'pow' else 'docker-compose-py.yml'
    compose_path = Path.cwd().parent.absolute()
    start_compose(compose_path, compose_file)
    now = datetime.now()
    filename_base = f'data_{now.strftime("%d_%m_%H-%M")}.xlsx'
    cont_prefix = 'pow_' if consensus == 'pow' else 'evoting_'
    cont_suffix = '' if consensus == 'pow' else '_1'

    queries = [
        ('container_memory_usage_bytes', 'memory'), 
        ('container_cpu_load_average_10s', 'cpu_10s'),
        ('container_network_transmit_bytes_total', 'network_tx'),
        ('container_network_receive_bytes_total', 'network_rcv'),
    ]

    notebooks = {}
    rounds_data = []

    for _, label in queries:
        notebooks[label] = xlsxwriter.Workbook(filename=consensus+'_'+label+'_'+filename_base)

    for round in range(rounds):
        xlsx_data = {}
        restart_containers(Path.cwd().parent.absolute(), 'docker-compose-py.yml')
        
        print('[INFO] starting round', round)
        
        start = int(datetime.now().timestamp())
        total_time: float = test_performance(transactions=transactions, node_address=node_address, consensus=consensus)
        end = int(datetime.now().timestamp())
        
        print('[INFO] round', round, 'done')
        
        for i in range(1, number_of_nodes+1):
            xlsx_data[f'node-{i}'] = {}
            cont_name = f'{cont_prefix}node-{i}{cont_suffix}'
            for query, label in queries:
                xlsx_data[f'node-{i}'][label] = query_usage(query, cont_name, start, end)

        dump_data_to_xlsx(notebooks, xlsx_data)
        rounds_data.append({
            'round_number': round+1,
            'transactions': transactions,
            'total_time': total_time,
        })

    for _, nb in notebooks.items():
        nb.close()

    rounds_workbook = Workbook(consensus+'_rounds_'+filename_base)
    dump_round_data(rounds_workbook, rounds_data)
    rounds_workbook.close()
    close_compose(compose_path, compose_file)
    

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Blockchain performance test suite')
    parser.add_argument('-t', '--transactions', type=int, help='number of transactions submitted per round', default=1000)
    parser.add_argument('-r', '--rounds', type=int, help='number of testing rounds', default=1)
    
    args = parser.parse_args()
    NUMBER_OF_ROUNDS = args.rounds
    NUMBER_OF_TRANSACTIONS = args.transactions
    start_tests_for_consensus(consensus='pbft', transactions=NUMBER_OF_TRANSACTIONS, rounds=NUMBER_OF_ROUNDS, node_address='http://localhost:2001', number_of_nodes=10)
    start_tests_for_consensus(consensus='pow',  transactions=NUMBER_OF_TRANSACTIONS, rounds=NUMBER_OF_ROUNDS, node_address='http://localhost:1337', number_of_nodes=3)
