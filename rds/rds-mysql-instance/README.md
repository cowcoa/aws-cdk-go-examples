## AWS PI to Grafana
Demonstrate how to sync AWS RDS/Aurora's Performance Insights metrics to Managed Grafana.<br />
Since Performance Insights only measures one DB instance at a time, so does this example, which can only sync PI metrics for a specified DB instance to the Grafana dashboard. But you can extend this example to the cluster level by yourself.

![alt text](../master/assets/aws-pi-to-grafana.png?raw=true)
![alt text](../master/assets/ScreenShot_01.png?raw=true)
![alt text](../master/assets/ScreenShot_02.png?raw=true)

## Supported Metrics 

grafana.db_status:

| Metric | Unit | Description |
| ------ | ------ | ------ |
| db.SQL.Com_analyze | Queries per second | Number of ANALYZE commands executed |
| db.SQL.Com_optimize | Queries per second | Number of OPTIMIZE commands executed |
| db.SQL.Com_select | Queries per second | Number of SELECT commands executed |
| db.SQL.Innodb_rows_inserted | Rows per second | Total rows inserted by InnoDB |
| db.SQL.Innodb_rows_deleted | Rows per second | Total rows deleted by InnoDB |
| db.SQL.Innodb_rows_updated | Rows per second | Total rows updated by InnoDB |
| db.SQL.Innodb_rows_read | Rows per second | Total rows read by InnoDB |
| db.SQL.Questions | Queries per second | The number of statements executed by the server. This includes only statements sent to the server by clients and not statements executed within stored programs |
| db.SQL.Queries | Queries per second | The number of statements executed by the server. This variable includes statements executed within stored programs |
| db.SQL.Select_full_join | Queries per second | The number of joins that perform table scans because they do not use indexes. If this value is not 0 you should carefully check the indexes of your tables |
| db.SQL.Select_full_range_join | Queries per second | The number of joins that used a range search on a reference table |
| db.SQL.Select_range | Queries per second | The number of joins that used ranges on the first table. This is normally not a critical issue even if the value is quite large |
| db.SQL.Select_range_check | Queries per second | The number of joins without keys that check for key usage after each row. If this is not 0 you should carefully check the indexes of your tables |
| db.SQL.Select_scan | Queries per second | The number of joins that did a full scan of the first table |
| db.SQL.Slow_queries | Queries per second | The number of queries that have taken more than long_query_time seconds. This counter increments regardless of whether the slow query log is enabled |
| db.SQL.Sort_merge_passes | Queries per second | The number of merge passes that the sort algorithm has had to do. If this value is large you should consider increasing the value of the sort_buffer_size system variable |
| db.SQL.Sort_range | Queries per second | The number of sorts that were done using ranges |
| db.SQL.Sort_rows | Queries per second | The number of sorted rows |
| db.SQL.Sort_scan | Queries per second | The number of sorts that were done by scanning the table |
| db.Locks.Innodb_row_lock_time | Milliseconds | The total time spent in acquiring row locks for InnoDB tables in milliseconds. |
| db.Locks.innodb_row_lock_waits | Transactions | The number of times operations on InnoDB tables had to wait for a row lock |
| db.Locks.innodb_deadlocks | Deadlocks per minute | Number of deadlocks |
| db.Locks.innodb_lock_timeouts | Timeouts | Number of InnoDB lock timeouts |
| db.Locks.Table_locks_immediate | Requests per second | The number of times that a request for a table lock could be granted immediately |
| db.Locks.Table_locks_waited | Requests per second | The number of times that a request for a table lock could not be granted immediately and a wait was needed |
| db.Users.Connections | Connections | The number of connection attempts to the MySQL server |
| db.Users.Aborted_clients | Connections | The number of connections that were aborted because the client died without closing the connection properly |
| db.Users.Aborted_connects | Connections | The number of failed attempts to connect to the MySQL server |
| db.Users.Threads_running | Connections | The number of threads that are not sleeping |
| db.Users.Threads_created | Connections | The number of threads created to handle connections |
| db.Users.Threads_connected | Connections | The number of currently open connections |
| db.IO.Innodb_pages_written | Pages per second | The number of pages written by operations on InnoDB tables |
| db.IO.Innodb_data_writes | Operations per second | The number InnoDB data write operations |
| db.IO.Innodb_log_writes | Operations per second | The number of physical writes to the InnoDB redo log |
| db.IO.Innodb_log_write_requests | Operations per second | The Number of requests to write to the InnoDB redo log |
| db.IO.Innodb_dblwr_writes | Operations per second | The number of writes done to the InnoDB double write buffer |
| db.Temp.Created_tmp_disk_tables | Tables per second | The number of internal on-disk temporary tables created by the server while executing statements |
| db.Temp.Created_tmp_tables | Tables per second | The number of internal temporary tables created by the server while executing statements |
| db.Transactions.active_transactions | Transactions | Number of Active transactions |
| db.Cache.Innodb_buffer_pool_reads | Pages per second | The number of logical reads that InnoDB could not satisfy from the buffer pool and had to read directly from disk |
| db.Cache.Innodb_buffer_pool_read_requests | Pages per second | The number of logical read requests |
| db.Cache.Innodb_buffer_pool_pages_data | Pages | The number of pages in the InnoDB buffer pool containing data. The number includes both dirty and clean pages |
| db.Cache.Innodb_buffer_pool_pages_total | Pages | The total size of the InnoDB buffer pool in pages |
| db.Cache.Opened_tables | Tables | The number of tables that have been opened. If Opened_tables is big your table_open_cache value is probably too small |
| db.Cache.Opened_table_definitions | Tables | The number of .frm files that have been cached |
| db.Transactions.trx_rseg_history_len | Length | Length of the TRX_RSEG_HISTORY list |
| db.Cache.innoDB_buffer_pool_hits | Pages per second | The number of reads that InnoDB could satisfy from the buffer pool |
| db.Cache.innoDB_buffer_pool_hit_rate | Percentage | The percentage of reads that InnoDB could satisfy from the buffer pool |
| db.Cache.innoDB_buffer_pool_usage | Percentage | The percentage of the InnoDB buffer pool that contains data (pages) |
| db.IO.innoDB_datafile_writes_to_disk | Operations per second | Number of InnoDB datafile writes to disk excluding doublewrite and redo logging write operations |
| db.SQL.innodb_rows_changed | Rows per second | Total InnoDB row operations |

For more metrics information, please refer to [db.metrics.txt](assets/db.metrics.txt), [os.metrics.txt](assets/os.metrics.txt) and [db.sql_tokenized.stats.metrics.txt](assets/db.sql_tokenized.stats.metrics.txt).

## Prerequisites
1. Install and configure AWS CLI Version 2 environment:<br />
   [Installation] - Installing or updating the latest version of the AWS CLI v2.<br />
   [Configuration] - Configure basic settings that AWS CLI uses to interact with AWS.<br />
   NOTE: Make sure your IAM User/Role has sufficient permissions.
2. Install Node Version Manager:<br />
   [Install NVM] - Install NVM and configure your environment according to this document.
3. Install Node.js:<br />
    ```sh
    nvm install 16.3.0
    ```
4. Install AWS CDK Toolkit:
    ```sh
    npm install -g aws-cdk
    ```
5. Install Golang:<br />
   [Download and Install] - Download and install Go quickly with the steps described here.
6. Install Docker:<br />
   [Install Docker Engine] - The installation section shows you how to install Docker on a variety of platforms.
7. Make sure you also have GNU Make, jq installed:<br />
    ```sh
    sudo yum install -y make
    sudo yum install -y jq
    ```

## Configuration

You can edit the cdk.json file to modify the deployment configuration.

| Key | Example Value | Description |
| ------ | ------ | ------ |
| stackName | PI2Grafana | CloudFormation stack name. It's difficult for jq to process JSON keys containing '-', so avoid naming your stack that way. |
| deploymentRegion | ap-northeast-1 | CloudFormation stack deployment region. If the value is empty, the default is the same as the region where deploy is executed. |
| dbInstanceName | my-rds-db-instance | RDS/Aurora DB instance name. This instance is your monitoring target, so make sure it already exists before deploying this example. |

## Deployment
1. Run the following command to deploy AWS infra and code by CDK Toolkit:<br />
     ```sh
     cdk-cli-wrapper-dev.sh deploy
     ```
   You can also clean up the deployment by running command:<br />
     ```sh
     cdk-cli-wrapper-dev.sh destroy
     ```
2. Run the following command to create AWS Managed Grafana workspace:<br />
     ```sh
     grafana/create-grafana-workspace.sh create
     ```
   This command will also save the Grafana workspace information to grafana/grafana-workspace-info.json file.<br />
   You can also clean up the Grafana workspace by running command:<br />
     ```sh
     grafana/create-grafana-workspace.sh delete
     ```
3. Run the following command to create default MySQL DataSource in Grafana workspace:<br />
     ```sh
     grafana/create-grafana-datasource.sh
     ```
   This command will also save the MySQL database information to grafana/mysql-datasource-info.json file.
4. Sign in to your AWS Web Console.
5. You can refer to [this blog](https://aws.amazon.com/blogs/security/how-to-create-and-manage-users-within-aws-sso/) to create AWS SSO User.
6. Find the Grafana workspace you just created and [add the AWS SSO User](https://docs.aws.amazon.com/grafana/latest/userguide/AMG-manage-users-and-groups-AMG.html) as the ADMIN.
7. Sign in to the Grafana workspace and find the MySQL DataSource you just created, fill in the password(according to the grafana/mysql-datasource-info.json file), and click "Save & Test".
8. Import Grafana dashboard by upload grafana/grafana-dashboard.json file.

[Installation]: <https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html>
[Configuration]: <https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html>
[Install NVM]: <https://github.com/nvm-sh/nvm#install--update-script>
[Download and Install]: <https://go.dev/doc/install>
[Install Docker Engine]: <https://docs.docker.com/engine/install/>
