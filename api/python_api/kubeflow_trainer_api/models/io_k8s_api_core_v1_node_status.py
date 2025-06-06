# coding: utf-8

"""
    Kubeflow Trainer OpenAPI Spec

    No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

    The version of the OpenAPI document: unversioned
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


from __future__ import annotations
import pprint
import re  # noqa: F401
import json

from pydantic import BaseModel, ConfigDict, Field, StrictStr, field_validator
from typing import Any, ClassVar, Dict, List, Optional
from kubeflow_trainer_api.models.io_k8s_api_core_v1_attached_volume import IoK8sApiCoreV1AttachedVolume
from kubeflow_trainer_api.models.io_k8s_api_core_v1_container_image import IoK8sApiCoreV1ContainerImage
from kubeflow_trainer_api.models.io_k8s_api_core_v1_node_address import IoK8sApiCoreV1NodeAddress
from kubeflow_trainer_api.models.io_k8s_api_core_v1_node_condition import IoK8sApiCoreV1NodeCondition
from kubeflow_trainer_api.models.io_k8s_api_core_v1_node_config_status import IoK8sApiCoreV1NodeConfigStatus
from kubeflow_trainer_api.models.io_k8s_api_core_v1_node_daemon_endpoints import IoK8sApiCoreV1NodeDaemonEndpoints
from kubeflow_trainer_api.models.io_k8s_api_core_v1_node_features import IoK8sApiCoreV1NodeFeatures
from kubeflow_trainer_api.models.io_k8s_api_core_v1_node_runtime_handler import IoK8sApiCoreV1NodeRuntimeHandler
from kubeflow_trainer_api.models.io_k8s_api_core_v1_node_system_info import IoK8sApiCoreV1NodeSystemInfo
from kubeflow_trainer_api.models.io_k8s_apimachinery_pkg_api_resource_quantity import IoK8sApimachineryPkgApiResourceQuantity
from typing import Optional, Set
from typing_extensions import Self

class IoK8sApiCoreV1NodeStatus(BaseModel):
    """
    NodeStatus is information about the current status of a node.
    """ # noqa: E501
    addresses: Optional[List[IoK8sApiCoreV1NodeAddress]] = Field(default=None, description="List of addresses reachable to the node. Queried from cloud provider, if available. More info: https://kubernetes.io/docs/reference/node/node-status/#addresses Note: This field is declared as mergeable, but the merge key is not sufficiently unique, which can cause data corruption when it is merged. Callers should instead use a full-replacement patch. See https://pr.k8s.io/79391 for an example. Consumers should assume that addresses can change during the lifetime of a Node. However, there are some exceptions where this may not be possible, such as Pods that inherit a Node's address in its own status or consumers of the downward API (status.hostIP).")
    allocatable: Optional[Dict[str, IoK8sApimachineryPkgApiResourceQuantity]] = Field(default=None, description="Allocatable represents the resources of a node that are available for scheduling. Defaults to Capacity.")
    capacity: Optional[Dict[str, IoK8sApimachineryPkgApiResourceQuantity]] = Field(default=None, description="Capacity represents the total resources of a node. More info: https://kubernetes.io/docs/reference/node/node-status/#capacity")
    conditions: Optional[List[IoK8sApiCoreV1NodeCondition]] = Field(default=None, description="Conditions is an array of current observed node conditions. More info: https://kubernetes.io/docs/reference/node/node-status/#condition")
    config: Optional[IoK8sApiCoreV1NodeConfigStatus] = Field(default=None, description="Status of the config assigned to the node via the dynamic Kubelet config feature.")
    daemon_endpoints: Optional[IoK8sApiCoreV1NodeDaemonEndpoints] = Field(default=None, description="Endpoints of daemons running on the Node.", alias="daemonEndpoints")
    features: Optional[IoK8sApiCoreV1NodeFeatures] = Field(default=None, description="Features describes the set of features implemented by the CRI implementation.")
    images: Optional[List[IoK8sApiCoreV1ContainerImage]] = Field(default=None, description="List of container images on this node")
    node_info: Optional[IoK8sApiCoreV1NodeSystemInfo] = Field(default=None, description="Set of ids/uuids to uniquely identify the node. More info: https://kubernetes.io/docs/reference/node/node-status/#info", alias="nodeInfo")
    phase: Optional[StrictStr] = Field(default=None, description="NodePhase is the recently observed lifecycle phase of the node. More info: https://kubernetes.io/docs/concepts/nodes/node/#phase The field is never populated, and now is deprecated.  Possible enum values:  - `\"Pending\"` means the node has been created/added by the system, but not configured.  - `\"Running\"` means the node has been configured and has Kubernetes components running.  - `\"Terminated\"` means the node has been removed from the cluster.")
    runtime_handlers: Optional[List[IoK8sApiCoreV1NodeRuntimeHandler]] = Field(default=None, description="The available runtime handlers.", alias="runtimeHandlers")
    volumes_attached: Optional[List[IoK8sApiCoreV1AttachedVolume]] = Field(default=None, description="List of volumes that are attached to the node.", alias="volumesAttached")
    volumes_in_use: Optional[List[StrictStr]] = Field(default=None, description="List of attachable volumes in use (mounted) by the node.", alias="volumesInUse")
    __properties: ClassVar[List[str]] = ["addresses", "allocatable", "capacity", "conditions", "config", "daemonEndpoints", "features", "images", "nodeInfo", "phase", "runtimeHandlers", "volumesAttached", "volumesInUse"]

    @field_validator('phase')
    def phase_validate_enum(cls, value):
        """Validates the enum"""
        if value is None:
            return value

        if value not in set(['Pending', 'Running', 'Terminated']):
            raise ValueError("must be one of enum values ('Pending', 'Running', 'Terminated')")
        return value

    model_config = ConfigDict(
        populate_by_name=True,
        validate_assignment=True,
        protected_namespaces=(),
    )


    def to_str(self) -> str:
        """Returns the string representation of the model using alias"""
        return pprint.pformat(self.model_dump(by_alias=True))

    def to_json(self) -> str:
        """Returns the JSON representation of the model using alias"""
        # TODO: pydantic v2: use .model_dump_json(by_alias=True, exclude_unset=True) instead
        return json.dumps(self.to_dict())

    @classmethod
    def from_json(cls, json_str: str) -> Optional[Self]:
        """Create an instance of IoK8sApiCoreV1NodeStatus from a JSON string"""
        return cls.from_dict(json.loads(json_str))

    def to_dict(self) -> Dict[str, Any]:
        """Return the dictionary representation of the model using alias.

        This has the following differences from calling pydantic's
        `self.model_dump(by_alias=True)`:

        * `None` is only added to the output dict for nullable fields that
          were set at model initialization. Other fields with value `None`
          are ignored.
        """
        excluded_fields: Set[str] = set([
        ])

        _dict = self.model_dump(
            by_alias=True,
            exclude=excluded_fields,
            exclude_none=True,
        )
        # override the default output from pydantic by calling `to_dict()` of each item in addresses (list)
        _items = []
        if self.addresses:
            for _item_addresses in self.addresses:
                if _item_addresses:
                    _items.append(_item_addresses.to_dict())
            _dict['addresses'] = _items
        # override the default output from pydantic by calling `to_dict()` of each value in allocatable (dict)
        _field_dict = {}
        if self.allocatable:
            for _key_allocatable in self.allocatable:
                if self.allocatable[_key_allocatable]:
                    _field_dict[_key_allocatable] = self.allocatable[_key_allocatable].to_dict()
            _dict['allocatable'] = _field_dict
        # override the default output from pydantic by calling `to_dict()` of each value in capacity (dict)
        _field_dict = {}
        if self.capacity:
            for _key_capacity in self.capacity:
                if self.capacity[_key_capacity]:
                    _field_dict[_key_capacity] = self.capacity[_key_capacity].to_dict()
            _dict['capacity'] = _field_dict
        # override the default output from pydantic by calling `to_dict()` of each item in conditions (list)
        _items = []
        if self.conditions:
            for _item_conditions in self.conditions:
                if _item_conditions:
                    _items.append(_item_conditions.to_dict())
            _dict['conditions'] = _items
        # override the default output from pydantic by calling `to_dict()` of config
        if self.config:
            _dict['config'] = self.config.to_dict()
        # override the default output from pydantic by calling `to_dict()` of daemon_endpoints
        if self.daemon_endpoints:
            _dict['daemonEndpoints'] = self.daemon_endpoints.to_dict()
        # override the default output from pydantic by calling `to_dict()` of features
        if self.features:
            _dict['features'] = self.features.to_dict()
        # override the default output from pydantic by calling `to_dict()` of each item in images (list)
        _items = []
        if self.images:
            for _item_images in self.images:
                if _item_images:
                    _items.append(_item_images.to_dict())
            _dict['images'] = _items
        # override the default output from pydantic by calling `to_dict()` of node_info
        if self.node_info:
            _dict['nodeInfo'] = self.node_info.to_dict()
        # override the default output from pydantic by calling `to_dict()` of each item in runtime_handlers (list)
        _items = []
        if self.runtime_handlers:
            for _item_runtime_handlers in self.runtime_handlers:
                if _item_runtime_handlers:
                    _items.append(_item_runtime_handlers.to_dict())
            _dict['runtimeHandlers'] = _items
        # override the default output from pydantic by calling `to_dict()` of each item in volumes_attached (list)
        _items = []
        if self.volumes_attached:
            for _item_volumes_attached in self.volumes_attached:
                if _item_volumes_attached:
                    _items.append(_item_volumes_attached.to_dict())
            _dict['volumesAttached'] = _items
        return _dict

    @classmethod
    def from_dict(cls, obj: Optional[Dict[str, Any]]) -> Optional[Self]:
        """Create an instance of IoK8sApiCoreV1NodeStatus from a dict"""
        if obj is None:
            return None

        if not isinstance(obj, dict):
            return cls.model_validate(obj)

        _obj = cls.model_validate({
            "addresses": [IoK8sApiCoreV1NodeAddress.from_dict(_item) for _item in obj["addresses"]] if obj.get("addresses") is not None else None,
            "allocatable": dict(
                (_k, IoK8sApimachineryPkgApiResourceQuantity.from_dict(_v))
                for _k, _v in obj["allocatable"].items()
            )
            if obj.get("allocatable") is not None
            else None,
            "capacity": dict(
                (_k, IoK8sApimachineryPkgApiResourceQuantity.from_dict(_v))
                for _k, _v in obj["capacity"].items()
            )
            if obj.get("capacity") is not None
            else None,
            "conditions": [IoK8sApiCoreV1NodeCondition.from_dict(_item) for _item in obj["conditions"]] if obj.get("conditions") is not None else None,
            "config": IoK8sApiCoreV1NodeConfigStatus.from_dict(obj["config"]) if obj.get("config") is not None else None,
            "daemonEndpoints": IoK8sApiCoreV1NodeDaemonEndpoints.from_dict(obj["daemonEndpoints"]) if obj.get("daemonEndpoints") is not None else None,
            "features": IoK8sApiCoreV1NodeFeatures.from_dict(obj["features"]) if obj.get("features") is not None else None,
            "images": [IoK8sApiCoreV1ContainerImage.from_dict(_item) for _item in obj["images"]] if obj.get("images") is not None else None,
            "nodeInfo": IoK8sApiCoreV1NodeSystemInfo.from_dict(obj["nodeInfo"]) if obj.get("nodeInfo") is not None else None,
            "phase": obj.get("phase"),
            "runtimeHandlers": [IoK8sApiCoreV1NodeRuntimeHandler.from_dict(_item) for _item in obj["runtimeHandlers"]] if obj.get("runtimeHandlers") is not None else None,
            "volumesAttached": [IoK8sApiCoreV1AttachedVolume.from_dict(_item) for _item in obj["volumesAttached"]] if obj.get("volumesAttached") is not None else None,
            "volumesInUse": obj.get("volumesInUse")
        })
        return _obj


