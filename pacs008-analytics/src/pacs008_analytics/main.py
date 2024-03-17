from fastapi import FastAPI, Request
import pandas as pd
from fastapi.responses import HTMLResponse
import json
from collections import defaultdict




app = FastAPI()

analytics_results = {}  # Global variable to store the analytics results

@app.post("/pacs008-analytics")
async def pacs008_analytics(request: Request):
    pacs008_data = await request.json()
    print(f"the data is: {pacs008_data}")

    transactions = []

    for tx in pacs008_data['FIToFICstmrCdtTrf']['CdtTrfTxInf']:
        amt = tx['Amt']['InstdAmt']
        transaction = defaultdict(str)
        transaction['Amount'] = float(amt[''])
        transaction['Currency'] = amt['Ccy']
        transactions.append(transaction)

    pacs008_df = pd.DataFrame(transactions)

    total_amount = pacs008_df[pacs008_df['Currency'] == 'EUR']['Amount'].sum()
    currency_counts = pacs008_df['Currency'].value_counts()
    max_amount = pacs008_df[pacs008_df['Currency'] == 'EUR']['Amount'].max()
    min_amount = pacs008_df[pacs008_df['Currency'] == 'EUR']['Amount'].min()

    analytics_results['total_amount'] = total_amount
    analytics_results['currency_counts'] = currency_counts.to_dict()
    analytics_results['max_amount'] = max_amount
    analytics_results['min_amount'] = min_amount

    return analytics_results

@app.get("/", response_class=HTMLResponse)
def analytics_report():
    total_amount = analytics_results.get('total_amount', 0)
    currency_counts = analytics_results.get('currency_counts', {})
    max_amount = analytics_results.get('max_amount', 0)
    min_amount = analytics_results.get('min_amount', 0)

    html_content = f"""
    <html>
    <head>
    <title>PACS.008 Analytics</title>
    </head>
    <body>
    <h1>PACS.008 Analytics Results</h1>
    <p>Total Amount: {total_amount}</p>
    <p>Maximum Amount: {max_amount}</p>
    <p>Minimum Amount: {min_amount}</p>
    <h2>Currency Counts:</h2>
    <ul>
    {''.join(f'<li>{currency}: {count}</li>' for currency, count in currency_counts.items())}
    </ul>
    </body>
    </html>
    """
    return HTMLResponse(content=html_content, status_code=200)


