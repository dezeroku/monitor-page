# coding: utf-8

from __future__ import absolute_import
from datetime import date, datetime  # noqa: F401

from typing import List, Dict  # noqa: F401

from swagger_server.models.base_model_ import Model
from swagger_server.models.attachment import Attachment  # noqa: F401,E501
from swagger_server import util


class Data(Model):
    """NOTE: This class is auto generated by the swagger code generator program.

    Do not edit the class manually.
    """
    def __init__(self, url: str=None, attachments: List[Attachment]=None):  # noqa: E501
        """Data - a model defined in Swagger

        :param url: The url of this Data.  # noqa: E501
        :type url: str
        :param attachments: The attachments of this Data.  # noqa: E501
        :type attachments: List[Attachment]
        """
        self.swagger_types = {
            'url': str,
            'attachments': List[Attachment]
        }

        self.attribute_map = {
            'url': 'url',
            'attachments': 'attachments'
        }
        self._url = url
        self._attachments = attachments

    @classmethod
    def from_dict(cls, dikt) -> 'Data':
        """Returns the dict as a model

        :param dikt: A dict.
        :type: dict
        :return: The Data of this Data.  # noqa: E501
        :rtype: Data
        """
        return util.deserialize_model(dikt, cls)

    @property
    def url(self) -> str:
        """Gets the url of this Data.


        :return: The url of this Data.
        :rtype: str
        """
        return self._url

    @url.setter
    def url(self, url: str):
        """Sets the url of this Data.


        :param url: The url of this Data.
        :type url: str
        """
        if url is None:
            raise ValueError("Invalid value for `url`, must not be `None`")  # noqa: E501

        self._url = url

    @property
    def attachments(self) -> List[Attachment]:
        """Gets the attachments of this Data.


        :return: The attachments of this Data.
        :rtype: List[Attachment]
        """
        return self._attachments

    @attachments.setter
    def attachments(self, attachments: List[Attachment]):
        """Sets the attachments of this Data.


        :param attachments: The attachments of this Data.
        :type attachments: List[Attachment]
        """
        if attachments is None:
            raise ValueError("Invalid value for `attachments`, must not be `None`")  # noqa: E501

        self._attachments = attachments
