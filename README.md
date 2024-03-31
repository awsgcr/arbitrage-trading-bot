## Arbitrage Trading Bot

This bot is designed to take advantage of arbitrage opportunities in the cryptocurrency market.
It is designed to be run on a server and will continuously monitor the market for arbitrage opportunities.
When an opportunity is found, the bot will execute trades to take advantage of the price difference.

![architecture.png](docs%2Farchitecture.png)

### Installation

1. Clone the repository
2. Install the required packages using ```go mod tidy```
3. Create a conf/secrets.conf.yml file in the ~/.coin_labor/ directory and add the following environment variables:
    - key: Your API key for the exchange
    - secret: Your API secret for the exchange

Example:
create file ~/.coin_labor/conf/secrets.conf.yml abd add the following

```
binance:
  key: "your binance key"
  secret: "your binance secret"

mexc:
  key: "your mexc key"
  secret: "your mexc secret"
```

4. Enable Alerting if needed, update token in the conf/dev.ini or conf/prod.ini file

5. Build the bot using ```go run build.go coin_labor```
6. Run the bot using below commands for different environments
    - For development environment
    ```
    ./bin/linux-amd64/coin_labor
    ```
    - For production environment
    ```
    ./bin/linux-amd64/coin_labor -config=conf/prod.ini
    ```

### Project Structure

- /conf - Contains the configuration files
- /core - Contains the basic components, settings and utilities
- /pkg - Contains the main logic of the bot
    - /cmd - Contains the main entry point of the bot
    - /components - Contains the components for each exchange
    - /plugins - Contains the plugins for each exchange
    - /services - Monitor the market for arbitrage opportunities

### Project Plan

* [x] Connect to different exchanges, Binance, OKX, MEXC
* [x] Fetch the order book from each exchanges
* [x] Calculate the spread between these exchanges
* [x] Monitor the market for arbitrage opportunities and send the data to Amazon Managed Service for Prometheus
* [ ] execute trades on an exchange