Taken from
 https://www.dell.com/support/manuals/us/en/04/system-update/dsu_ug_1.8_revamp/dsu-return-codes?guid=guid-a413b447-0dd2-45fb-a60c-7a472e353e30&lang=en-us


DSU Return Codes
Number	Return Codes	Description of Return Codes
0	Success	Any successful operation performed by DSU.
1	Failure	Any failure in operation performed by DSU.
2	Insufficient Privileges	DSU not executed using ROOT privilege..
3	Invalid Log File	Failure in opening a log file or invalid log location.
4	Invalid Log Level	Invalid log level set by user.
6	Invalid Command Line Option	Invalid combination of DSU options used. For example, –destination–destination type and –non-interactive–non-interactive cannot be used simultaneously.
7	Unknown Option	Incorrect option provided.
8	Reboot Required	Reboot is required for the update to be completed successfully.
12	Authentication failure	When the provided credentials during the network share are incorrect, the following return code is displayed
13	Invalid Source Config (Configuration)	Values provided for source location or source type is invalid.
14	Invalid Inventory	Errors related to Inventory such as filename not present in the location or failed parsing inventory.
15	Invalid Category	Category value (for example: BI) may not exist, DSU returns Invalid Category
17	Invalid Config (Configuration) File	Configuration file location is invalid or failure in parsing it.
19	Invalid IC Location	Invalid Location of inventory collector.
20	Invalid Component Type	Any component type other than the specified type, displays invalid component type
21	Invalid Destination	Destination directory location is invalid.
22	Invalid Destination Type	Destination type is not ISO or CBD.
24	Update Failure	Failure in applying updates.
25	Update Partial Failure	Out-of-date updates are selected.
26	Update Partial Failure And Reboot Required	Out-of-date updates are selected. For successful updates, reboot is required.
27	Destination not reachable	Unable to connect to the remote machine
28	Connection access denied	Privilege restriction
29	Connection invalid session	Abrupt closure of the session
30	Connection Time out	Connection to the system timed out
31	Connection unsupported protocol	Invalid protocols provided during the connection to remote system or target
32	Connection terminated	Connection to the system terminated
33	Execution permission denied	Restricted privilege
34	No Applicable Updates Found	There are no updates found which can be applied.
35	Remote Partial Failure	Some remote systems has failure some maybe successful.
36	Remote Failure	All the remote systems has failure.
37	IC Signature Download Failure	Unable to get the signature file for IC.
40	Public Key Not Found	The signature verification failed due to public keys are not imported on system.
41	No Progress available	Progress report not available