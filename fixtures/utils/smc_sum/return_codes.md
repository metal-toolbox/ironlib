Exit Code Number Description
0 Successful
Others Failed
GROUP1 (1~30) Command line parsing check failed

1 GetOpt unexpected option code
2 Unknown option
3 Missing argument
4 No host IP/user/password
5 Missing option
6 Unknown command
7 Option conflict
8 Can not open file
9 File already exists
10 Host is unknown
11 Invalid command line data
12 Function access denied

GROUP2 (31~59) Resource management error
31 File management error
32 Thread management error
33 TCP connection error
34 UDP connection error
35 Program interrupted and terminated
36 Required device does not exist
Supermicro Update Manager User’s Guide 134
37 Required device does not work
38 Function is not supported

GROUP3 (60~79) File parsing errors
60 Invalid BIOS configuration file
61 Utility internal error
62 Invalid firmware image file
63 Invalid firmware flash ROM
64 Invalid DMI information from BIOS
65 Invalid DMI information text file
66 Invalid DMI command line format
67 Invalid system list file
68 Invalid BMC configuration text file
69 Invalid asset information
70 Invalid CMM configuration text file
71 Invalid RAID configuration file
72 Invalid PCH asset information file format
73 Invalid full SMBIOS file format
74 Invalid VPD file format
75 Invalid BIOS internal file

GROUP4 (80~99) IPMI operation errors
80 Node Product key is not activated
81 Internal communication error
82 Board information mismatch
83 Does not support OOB
Supermicro Update Manager User’s Guide 135
84 Does not support get file
85 File is not available for download
86 Required tool does not exist
87 IPMI standard error

GROUP5 (100~119) In-band operation errors
100 Cannot open driver
101 Driver input/output control failed
102 Driver report: ****execution of command failed****
103 BIOS does not support this in-band command
104 Driver report: ****file size out of range****
105 Cannot load driver
106 Driver is busy. Please try again later
107 ROM chip is occupied. Please try again later
108 Kernel module verification error

GROUP6 (120~199) IPMI communication errors
144 IPMI undefined error
145 IPMI connect failed
146 IPMI login failed
147 IPMI execution parameter validation failed
148 IPMI execution exception occurred
149 IPMI execution failed
150 IPMI execution exception on slave CMM or unavailable
151 IPMI execution exception on module not present
152 IPMI execution only for CMM connected