# Overview

This section describes how to deploy Pachyderm in a production environment.
Additionally, you will find information about basic Pachyderm operations,
including upgrading to minor and major versions, autoscaling...

Before you start... The following high-level architecture diagram lays out Pachyderm's main components. It might help you build a quick mental model of Pachyderm.
![Operator High Level Arch](./images/arch_diagram_high_level.svg)

!!! Attention 
    We are now shipping Pachyderm with an **embedded proxy** 
    allowing your cluster to expose one single port externally. This deployment setup is optional.
    
    If you choose to deploy Pachyderm with a Proxy, check out our new recommended architecture and [deployment instructions](../deploy-manage/deploy/deploy-w-proxy/). 

<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Deploy Pachyderm &nbsp;&nbsp; &nbsp;<i class="fa fa-laptop"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
       Learn how to deploy Pachyderm in your cloud environment, or locally.
      </div>
      <div class="mdl-card__actions mdl-card--border">
          <ul>
            <li><a href="deploy/" class="md-typeset md-link">
            Deploy Pachyderm
            </a>
            </li>
          </ul>
      </div>
    </div>
  </div>
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Lifecycle Management &nbsp;&nbsp;&nbsp;<i class="fa fa-cogs"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Learn how to upgrade, backup, restore, and
        perform other Pachyderm management operations.
      </div>
      <div class="mdl-card__actions mdl-card--border">
          <ul>
            <li><a href="manage/" class="md-typeset md-link">
            Manage Pachyderm
           </a>
          </li>
       </div>
     </div>
  </div>
</div>
