import requests


def body(consensus: str, i: int, toId = None):
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

r = requests.post('http://localhost:2001/new-request', json=body('pbft', 1, 'sdfsdfsdf'))

print(r.json())