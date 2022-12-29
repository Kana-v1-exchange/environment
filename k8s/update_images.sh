docker build -t msolomodenko/exchange_calculator https://github.com/Kana-v1-exchange/calculator.git &
docker build -t msolomodenko/exchange_dashboard https://github.com/Kana-v1-exchange/dashboard.git &
docker build -t msolomodenko/exchange_frontend https://github.com/Kana-v1-exchange/frontend.git & 

wait 

docker push msolomodenko/exchange_calculator &
docker push msolomodenko/exchange_dashboard &
docker push msolomodenko/exchange_frontend