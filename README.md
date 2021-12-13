# ARL - API Rate Limiter

### Description:

Application designed to rate limit the api requests

### Tech specs:

Algorithm: `rolling window`

Storage: `Redis`

Minimum window size: `1 second`

Maximum window size: `1 minute`

### Requirements:

- Key size ~ 16 Bytes
- Epoch timestamp ~ 4 Bytes
- Counter ~ 4 Bytes

All data would store as hash table lets assume  20 Bytes extra for hashtable internals

System store each counter for each second in minute separated.

For each user we need a would use Redis hash lets assume 20 Bytes extra.

Total one user data weight is: `16 + (4 + 4 + 20) * 60 + 20 = 1.7 Kb`

If we need to track one million users at any time, total memory we would need would be 1.7 GB: `1.7  * 1 million = 1.7 GB`