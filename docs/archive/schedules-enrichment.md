We have a Schedules section where I can configure cron jobs that prompt the agent to check all connections of our app in intervals to check for smth.

Can you see that too?

It looks like this is a cron job for the Agent.

However, what if I've got connections that don't continuously push through logs into Heimdall? Meaning, what if I've got a connection that I have to proactively fetch logs from? 

Would/could that ever be the case? i.e. connectors that stream continuously vs connectors that could be log-requested ondemand?

If so, I'm thinking perhaps we should also be able to add Cron Jobs that simply fetch logs?

Dyou get what I'm describing/getting at?