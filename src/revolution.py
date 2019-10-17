# -*- coding: utf-8 -*-
#
from rest_framework import serializers

from django.utils.translation import ugettext_lazy as _

from orgs.mixins import OrgResourceSerializerMixin
from common.mixins import BulkSerializerMixin
from common.serializers import AdaptedBulkListSerializer
from ..models import Version

__all__ = [
    'VersionSerializer',
]


# 定义序列化类
class VersionSerializer(BulkSerializerMixin, serializers.ModelSerializer):
    """
    发版 历史
    """

    class Meta:
        model = Version
        list_serializer_class = AdaptedBulkListSerializer
        # 配置要序列化的字段
        fields = [
            'id', 'name',
            'is_active', 'date_created', 'projects', 'which_harbor'
        ]
        read_only_fields = (
            'date_created',
        )
        extra_kwargs = {
            'projects_id': {'label': _('Projects')}
        }

    @classmethod
    def setup_eager_loading(cls, queryset):
        """ Perform necessary eager loading of data. """
        # queryset = queryset.prefetch_related('labels', 'nodes')\
        #     .select_related('admin_user')
        return queryset

    def get_field_names(self, declared_fields, info):
        fields = super().get_field_names(declared_fields, info)
        fields.extend([
            'projectsname', 'projectsaddr'
        ])
        return fields


# class AssetGrantedSerializer(serializers.ModelSerializer):
#     """
#     被授权资产的数据结构
#     """
#     system_users_granted = AssetSystemUserSerializer(many=True, read_only=True)
#     system_users_join = serializers.SerializerMethodField()
#     # nodes = NodeTMPSerializer(many=True, read_only=True)
#
#     class Meta:
#         model = Images
#         fields = (
#             "id", "hostname", "ip", "port", "system_users_granted",
#             "is_active", "system_users_join", "os", 'domain',
#             "platform", "comment", "protocol", "org_id", "org_name",
#         )
#
#     @staticmethod
#     def get_system_users_join(obj):
#         system_users = [s.username for s in obj.system_users_granted]
#         return ', '.join(system_users)

#
# class MyAssetGrantedSerializer(AssetGrantedSerializer):
#     """
#     普通用户获取授权的资产定义的数据结构
#     """
#
#     class Meta:
#         model = Images
#         fields = (
#             "id", "hostname", "system_users_granted",
#             "is_active", "system_users_join", "org_name",
#             "os", "platform", "comment", "org_id", "protocol"
#         )


class ImagesSimpleSerializer(serializers.ModelSerializer):
    class Meta:
        model = Version
        fields = ['id', 'name', 'which_harbor']
