import requests
import time

NUMBER_OF_TRANSACTIONS = 1000
NODE_ADDRESS = 'http://localhost:2001'
CONSENSUS = 'pbft'

def body(consensus, i):
    if consensus == 'pow':
        body = {
            'Token': f'token {i}',
            'ToId': f'transaction {i}',
        }
        return body
    elif consensus == 'pbft':
        body = {
                'transactions': [
                    {
                        'Token': f'token {i}',
                        'ToId': f'transaction {i}',
                    }
                ]
            }
        return body
    else:
        raise Exception('incorrect consensus protocol')


if __name__ == '__main__':
    start = time.time()
    url = f'{NODE_ADDRESS}/transaction/create' if CONSENSUS == 'pow' else f'{NODE_ADDRESS}/new-request'
    for i in range(NUMBER_OF_TRANSACTIONS):
        r = requests.post(url, json=body(CONSENSUS, i))
    end = time.time()

    print(f"test complete, {NUMBER_OF_TRANSACTIONS} transactions took {end-start} s")
