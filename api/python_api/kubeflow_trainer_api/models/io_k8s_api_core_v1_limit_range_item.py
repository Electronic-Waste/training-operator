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

from pydantic import BaseModel, ConfigDict, Field, StrictStr
from typing import Any, ClassVar, Dict, List, Optional
from kubeflow_trainer_api.models.io_k8s_apimachinery_pkg_api_resource_quantity import IoK8sApimachineryPkgApiResourceQuantity
from typing import Optional, Set
from typing_extensions import Self

class IoK8sApiCoreV1LimitRangeItem(BaseModel):
    """
    LimitRangeItem defines a min/max usage limit for any resource that matches on kind.
    """ # noqa: E501
    default: Optional[Dict[str, IoK8sApimachineryPkgApiResourceQuantity]] = Field(default=None, description="Default resource requirement limit value by resource name if resource limit is omitted.")
    default_request: Optional[Dict[str, IoK8sApimachineryPkgApiResourceQuantity]] = Field(default=None, description="DefaultRequest is the default resource requirement request value by resource name if resource request is omitted.", alias="defaultRequest")
    max: Optional[Dict[str, IoK8sApimachineryPkgApiResourceQuantity]] = Field(default=None, description="Max usage constraints on this kind by resource name.")
    max_limit_request_ratio: Optional[Dict[str, IoK8sApimachineryPkgApiResourceQuantity]] = Field(default=None, description="MaxLimitRequestRatio if specified, the named resource must have a request and limit that are both non-zero where limit divided by request is less than or equal to the enumerated value; this represents the max burst for the named resource.", alias="maxLimitRequestRatio")
    min: Optional[Dict[str, IoK8sApimachineryPkgApiResourceQuantity]] = Field(default=None, description="Min usage constraints on this kind by resource name.")
    type: StrictStr = Field(description="Type of resource that this limit applies to.")
    __properties: ClassVar[List[str]] = ["default", "defaultRequest", "max", "maxLimitRequestRatio", "min", "type"]

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
        """Create an instance of IoK8sApiCoreV1LimitRangeItem from a JSON string"""
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
        # override the default output from pydantic by calling `to_dict()` of each value in default (dict)
        _field_dict = {}
        if self.default:
            for _key_default in self.default:
                if self.default[_key_default]:
                    _field_dict[_key_default] = self.default[_key_default].to_dict()
            _dict['default'] = _field_dict
        # override the default output from pydantic by calling `to_dict()` of each value in default_request (dict)
        _field_dict = {}
        if self.default_request:
            for _key_default_request in self.default_request:
                if self.default_request[_key_default_request]:
                    _field_dict[_key_default_request] = self.default_request[_key_default_request].to_dict()
            _dict['defaultRequest'] = _field_dict
        # override the default output from pydantic by calling `to_dict()` of each value in max (dict)
        _field_dict = {}
        if self.max:
            for _key_max in self.max:
                if self.max[_key_max]:
                    _field_dict[_key_max] = self.max[_key_max].to_dict()
            _dict['max'] = _field_dict
        # override the default output from pydantic by calling `to_dict()` of each value in max_limit_request_ratio (dict)
        _field_dict = {}
        if self.max_limit_request_ratio:
            for _key_max_limit_request_ratio in self.max_limit_request_ratio:
                if self.max_limit_request_ratio[_key_max_limit_request_ratio]:
                    _field_dict[_key_max_limit_request_ratio] = self.max_limit_request_ratio[_key_max_limit_request_ratio].to_dict()
            _dict['maxLimitRequestRatio'] = _field_dict
        # override the default output from pydantic by calling `to_dict()` of each value in min (dict)
        _field_dict = {}
        if self.min:
            for _key_min in self.min:
                if self.min[_key_min]:
                    _field_dict[_key_min] = self.min[_key_min].to_dict()
            _dict['min'] = _field_dict
        return _dict

    @classmethod
    def from_dict(cls, obj: Optional[Dict[str, Any]]) -> Optional[Self]:
        """Create an instance of IoK8sApiCoreV1LimitRangeItem from a dict"""
        if obj is None:
            return None

        if not isinstance(obj, dict):
            return cls.model_validate(obj)

        _obj = cls.model_validate({
            "default": dict(
                (_k, IoK8sApimachineryPkgApiResourceQuantity.from_dict(_v))
                for _k, _v in obj["default"].items()
            )
            if obj.get("default") is not None
            else None,
            "defaultRequest": dict(
                (_k, IoK8sApimachineryPkgApiResourceQuantity.from_dict(_v))
                for _k, _v in obj["defaultRequest"].items()
            )
            if obj.get("defaultRequest") is not None
            else None,
            "max": dict(
                (_k, IoK8sApimachineryPkgApiResourceQuantity.from_dict(_v))
                for _k, _v in obj["max"].items()
            )
            if obj.get("max") is not None
            else None,
            "maxLimitRequestRatio": dict(
                (_k, IoK8sApimachineryPkgApiResourceQuantity.from_dict(_v))
                for _k, _v in obj["maxLimitRequestRatio"].items()
            )
            if obj.get("maxLimitRequestRatio") is not None
            else None,
            "min": dict(
                (_k, IoK8sApimachineryPkgApiResourceQuantity.from_dict(_v))
                for _k, _v in obj["min"].items()
            )
            if obj.get("min") is not None
            else None,
            "type": obj.get("type") if obj.get("type") is not None else ''
        })
        return _obj


