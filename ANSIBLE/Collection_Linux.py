#!/usr/bin/env python



"""

1、本脚本采集服务器相关信息，ansible版本2.8.2
2、对ansible二次开发
3、详情参考ansible官网，2.8.x版本的官方api有问题，请自行修改

"""





import json
import shutil
from ansible.module_utils.common.collections import ImmutableDict
from ansible.parsing.dataloader import DataLoader
from ansible.vars.manager import VariableManager
from ansible.inventory.manager import InventoryManager
from ansible.playbook.play import Play
from ansible.executor.task_queue_manager import TaskQueueManager
from ansible.plugins.callback import CallbackBase
from ansible import context
import ansible.constants as C


class ResultCallback(CallbackBase):
    """A sample callback plugin used for performing an action as results come in

    If you want to collect all results into a single object for processing at
    the end of the execution, look into utilizing the ``json`` callback plugin
    or writing your own custom callback plugin
    """

    def v2_runner_on_ok(self, result, **kwargs):
        """Print a json representation of the result

        This method could store the result in an instance attribute for retrieval later
        """
        host = result._host
        # print(json.dumps({host.name: result._result}, indent=4))
        data = {
            "hostname": result._result["ansible_facts"]["ansible_nodename"],
            "system_vendor": result._result["ansible_facts"]["ansible_system_vendor"],
            "description": result._result["ansible_facts"]["ansible_lsb"]["description"],
            "ip": result._result["ansible_facts"]["ansible_all_ipv4_addresses"],
            "dns": result._result["ansible_facts"]["ansible_dns"],
            "version": result._result["ansible_facts"]["ansible_product_name"],
            "cpu": result._result["ansible_facts"]["ansible_processor_vcpus"],
            "kernel": result._result["ansible_facts"]["ansible_kernel"],
            "processor_cores": result._result["ansible_facts"]["ansible_processor_cores"],
            "processor_threads_per_core": result._result["ansible_facts"]["ansible_processor_threads_per_core"]
        }
        print(json.dumps({host.name: data}, indent=4))


# since the API is constructed for CLI it expects certain options to always be set in the context object
# context.CLIARGS = ImmutableDict(connection='local', module_path=['/to/mymodules'], forks=10, become=None,
#                                become_method=None, become_user=None, check=False, diff=False)

context.CLIARGS = ImmutableDict(connection='ssh', module_path=None, become=None, become_method=None, forks=10,
                                start_at_task=None, become_user=None, check=False, diff=False, syntax=None,
                                ansible_cfg=None, verbosity=3)

# initialize needed objects
loader = DataLoader()  # Takes care of finding and reading yaml, json and ini files
passwords = dict(vault_pass='')

# Instantiate our ResultCallback for handling results as they come in. Ansible expects this to be one of its main display outlets
results_callback = ResultCallback()

# create inventory, use path to host config file as source or hosts in a comma separated string
inventory = InventoryManager(loader=loader, sources='/ansible/sc.ini')

# variable manager takes care of merging all the different sources to give you a unified view of variables available in each context
variable_manager = VariableManager(loader=loader, inventory=inventory)

# create data structure that represents our play, including tasks, this is basically what our YAML loader does internally.
play_source = dict(
    name="Ansible Play",
    hosts='sc',
    gather_facts='no',
    tasks=[
        dict(action=dict(module='setup'), register=('shell_out')),
        # dict(action=dict(module='shell', args='pwd'), register='shell_out'),
        dict(action=dict(module='debug', args=dict(msg='{{shell_out.stdout}}')))

    ]
)

# Create play object, playbook objects use .load instead of init or new methods,
# this will also automatically create the task objects from the info provided in play_source
play = Play().load(play_source, variable_manager=variable_manager, loader=loader)

# Run it - instantiate task queue manager, which takes care of forking and setting up all objects to iterate over host list and tasks
tqm = None
try:
    tqm = TaskQueueManager(
        inventory=inventory,
        variable_manager=variable_manager,
        loader=loader,
        passwords=passwords,
        stdout_callback=results_callback,
        # Use our custom callback instead of the ``default`` callback plugin, which prints to stdout
    )
    result = tqm.run(play)  # most interesting data for a play is actually sent to the callback's methods
finally:
    # we always need to cleanup child procs and the structures we use to communicate with them
    if tqm is not None:
        tqm.cleanup()

    # Remove ansible tmpdir
    shutil.rmtree(C.DEFAULT_LOCAL_TMP, True)

