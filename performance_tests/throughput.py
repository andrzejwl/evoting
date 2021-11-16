import requests
import time

NUMBER_OF_TRANSACTIONS = 600
NODE_ADDRESS = 'http://localhost:2001'
CONSENSUS = 'pbft'

def body(consensus, i, toId = None):
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


if __name__ == '__main__':
    parties = ["voting party 1", "voting party 2"]
    start = time.time()
    url = f'{NODE_ADDRESS}/transaction/create' if CONSENSUS == 'pow' else f'{NODE_ADDRESS}/new-request'
    
    for token in parties:
        for i in range(NUMBER_OF_TRANSACTIONS//len(parties)):
            r = requests.post(url, json=body(CONSENSUS, i, token))
    end = time.time()

    print(f"test complete, {NUMBER_OF_TRANSACTIONS} transactions took {end-start} s")
