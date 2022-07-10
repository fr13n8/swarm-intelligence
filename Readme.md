#Swarm intelligence
>Swarm intelligence is defined as a collective behavior of a decentralized or self-organized system. These systems consist of numerous subjects with limited intelligence interacting with each other based on simple principles.

https://user-images.githubusercontent.com/57458390/178143547-60869f3a-1654-47ad-a0be-a7dd403e58b7.mp4

From [Wikipedia](https://en.wikipedia.org/wiki/Swarm_intelligence), the free encyclopedia:
>Swarm intelligence (SI) is the collective behavior of decentralized, self-organized systems, natural or artificial. The concept is employed in work on artificial intelligence. The expression was introduced by Gerardo Beni and Jing Wang in 1989, in the context of cellular robotic systems.
>The exact definition of swarm intelligence is still not formulated. In general, RI should be a multi-agent system that would have self-organizing behavior, which, in total, should exhibit some reasonable behavior.

Modeling the swarm behavior of agents who must look for resources and bring them home.Agents are blind, walk with crooked steps, but they can transmit signals and receive them from other such agents. Each signal contains information about the location of a particular agent in relation to resources or home. It sends this information in a signal to other agents, and then using this information, other agents can calculate the distance and direction from their location in relation to resources or home. To implement this functionality, you need to follow the rules:
- each agent has two direction counters in which the conditional number of steps to each point is recorded;
- for each step, the counters increase by one, regardless of direction;
- after **x** steps the agent transmits a signal with the value of the **first counter + the radius of y steps**, after **x/2** steps the value from the **second counter + the radius of y steps** (the radius is the maximum distance for which the signal is transmitted from the agent);
- the agent sets a goal to get to one of the points;
- at each step, the agent determines whether he is in the desired point, if his **(x, y)** coordinates coincide with the coordinates of the desired point, then he resets the counter of this point and turns **180°** and sets a new goal to be in the second point;
- if the agent stumbles upon a point that is not the desired one, then the agent continues its movement;
- if the agent is within the radius of the signal, then it compares the counter values ​​with the value of the received signal, if the value is less, than it updates the corresponding counter, and if this value corresponds to the target point, then the agent needs to move in the direction of the signal.

###Note!

**Agents should have a small percentage of route deviation, because if all agents move in a straight line, they will not be able to find the best route to the nearest point that is not on their path of movement.
The agents should have different movement speeds so that the movement is chaotic and the recon agents have the opportunity to deviate from the path.**

## About implementation
This is a running implementation of simple swarm intelligence written on go. \
For graphical implementation, the [ebiten](https://github.com/hajimehoshi/ebiten) 2d engine was used.
