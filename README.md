Sharding blockchain is a technology designed to improve the performance and scalability of traditional blockchain systems.
However, due to its design, communication between shards depends on shard leaders for transmitting information, while shard members are unable to detect communication activities between shards.
Consequently, Byzantine nodes can act as shard leaders, engaging in malicious behaviors to disrupt message transmission.

To address these issues, we propose the Cross-Shard Leader Re-election Protocol (CSLRE), which is based on the Two-Phase Atomic Commit Protocol (2PC).
CSLRE employs Byzantine Broadcast/Byzantine Agreement (BB/BA) for Byzantine fault tolerance to generate Cross-Shard Leader Re-election certificates, thereby reducing the impact of shard leaders on inter-shard communication.
It also uses Round-robin mechanism to facilitate leader re-election.
Moreover, we demonstrate that CSLRE maintains the security and liveness of sharding transactions while providing lower communication latency.
Finally, we conducted an experimental comparison between CSLRE and other cross-shard protocols.
The results indicate that CSLRE exhibits superior performance in reducing communication latency.
We deployed our implementation on a cloud instance from the Alibaba Cloud .

In our experiment, the main metrics were latency (quantified in seconds) and throughput (quantified in bytes processed per second).
We implement a two-stage atomic commit model under the standard synchronous model.
Shards mainly use BFTs based on the BB prototype to synchronize hotstuff (SYNC) and reputation-based state machine replication (RBSMR) for certificate information generation and final transaction confirmation.
We modified the core logic of SYNC and RBSMR by adding an cross-shard transaction model to ensure that messages sent from other shards are recognized and processed within the chip.
Shards use a 2PC to interact with each other.
CSVC,FS\_CSVC and CSLRE are three cross-shard lead conversion protocols.
For different cases, we use the following abbreviations:
When the intrashard consensus is SYNC, there is no cross-shard lead conversion protocols between shards.
In short, for SYNC, when CSVC is used in the 2PC, it is called SYNC-CSVC; When using FS_CSVC, it is called SYNC-FS; When CSLRE is used, it is called SYNC-CSLRE.
The same is true when the intrashard consensus is the RBSMR.
We use fmt to output content and log to record log information. We utilize net/http to simulate each of the different nodes.
Additionally, we use the sync package to achieve synchronization among multiple goroutines. Moreover, we have reused the intra-shard consensus operations from SYNC and RBSMR.
In our implementation, the transactions proposed by customers are all cross-shard transactions and are specific to T{(I1,I2)->(O1,O2,O3)}.
Shard_out_3 is the coordinator shard by default.
Other shards submit the certificates to Shard_out_3.
The model will only process the next transaction after executing one transaction.
All throughput and latency results are measured from clients of separate processes that run as separate processes on the same virtual machine as the shard members.
We ensure that the performance of sharding is not limited by the lack of transactions proposed by the client.

QUICK START
1. `go mod download` 
2. `/quicktest.sh`
3. In the corresponding file name, for example, nomal2PC. Start by opening the nomal2pc.go file in the appropriate folder. 
   Execute the file. You can see the execution, and the file is recorded as a file, adjusting the parameter values in `config/globals.go`.
   Other procedures are executed similarly.
4. In the `config/globals`. Adjust the **Delay , Round, Node, ShardNumber, GlobalPayload**.
   They represent the maximum round trip delay per round, the number of transaction rounds, the number of nodes in the shard, the number of shards, and the transaction size. 
   (Note: shardNumber is a new content and can only be controlled as 3-7) It cannot be added.
5. You can view all data records under the `/Dataforexpreiment/` file. The folder names named after different protocols.
   For example, 400T4N100D100R.txt under normal2PC/ represents the transaction size of 400, the maximum round-trip delay of 4,100 members per shard, and the result of 100 rounds of transactions. Others are similar.
