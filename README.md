# ARL - API Rate Limiter

### Description:

Application designed to rate limit the api requests

### Tech specs:

Algorithm: `rolling window`

Storage: `Redis`

Minimum window size: `1 second`

Maximum window size: `1 minute`

### Requirements:

- Key size ~ 8 Bytes
- Epoch timestamp ~ 4 Bytes
- Counter ~ 2 Bytes

All data would store as hash table lets assume  20 Bytes extra for hashtable internals

System store each counter for each second in minute separated.

For each user we need a would use Redis hash lets assume 20 Bytes extra.

Total one user data weight is: `8 + (4 + 2 + 20) * 60 + 20 = 1.6 Kb`

If we need to track one million users at any time, total memory we would need would be 1.6 GB: `1.6  * 1 million = 1.6 GB`