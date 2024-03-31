import json
import logging
import os
import socket
import time
import urllib
import uuid

import jwt
import pytest
import requests

LOGGER = logging.getLogger(__name__)


def wait_for_socket(host, port):
    retries = 10
    exception = None
    while retries > 0:
        try:
            socket.socket().connect((host, port))
            return
        except ConnectionRefusedError as e:
            exception = e
            print(f'Got ConnectionError for url {host}:{port}: {e} , retrying')
            retries -= 1
            time.sleep(2)
    raise exception


@pytest.fixture
def main_service_addr():
    addr = os.environ.get('MAIN_SERVICE_URL', 'http://127.0.0.1:8090')
    host = urllib.parse.urlparse(addr).hostname
    port = urllib.parse.urlparse(addr).port
    wait_for_socket(host, port)
    yield addr


@pytest.fixture
def jwt_private():
    path = os.environ.get('JWT_PRIVATE_KEY_FILE', '')
    with open(path, 'rb') as file:
        key = file.read()
    yield key


@pytest.fixture
def jwt_public():
    path = os.environ.get('JWT_PUBLIC_KEY_FILE', '')
    with open(path, 'rb') as file:
        key = file.read()
    yield key


def make_requests(method, addr, handle, params=None, data=None, cookies=None):
    if data is not None:
        data = json.dumps(data)
    req = requests.Request(
        method,
        addr +
        handle,
        params=params,
        data=data,
        cookies=cookies)
    prepared = req.prepare()
    LOGGER.info(f'>>> {prepared.method} {prepared.url}')
    if len(req.data) > 0:
        LOGGER.info(f'>>> {req.data}')
    if req.cookies is not None:
        LOGGER.info(f'>>> {req.cookies}')
    s = requests.Session()
    resp = s.send(prepared)
    LOGGER.info(f'<<< {resp.status_code}')
    if len(resp.content) > 0:
        LOGGER.info(f'<<< {resp.content}')
    if len(resp.cookies) > 0:
        LOGGER.info(f'<<< {resp.cookies}')
    return resp


def make_user(main_service_addr):
    login = str(uuid.uuid4())
    password = str(uuid.uuid4())
    r = make_requests(
        'POST',
        main_service_addr,
        '/signup',
        data={
            'login': login,
            'password': password})
    assert r.status_code == 200
    cookies = r.cookies.get_dict()
    return ((login, password), cookies)


@pytest.fixture
def user(main_service_addr):
    yield make_user(main_service_addr)


@pytest.fixture
def another_user(main_service_addr):
    yield make_user(main_service_addr)


def generate_jwt(private, id):
    return jwt.encode({'id': id}, private, 'RS256')


def parse_jwt(token, public):
    return jwt.decode(token, public, ['RS256'])


def generate_hs256_jwt(secret, id):
    return jwt.encode({'id': id}, secret, 'HS256')


class TestRSA:
    @staticmethod
    def test_private(jwt_private):
        generate_jwt(jwt_private, 12345)

    @staticmethod
    def test_public(jwt_private, jwt_public):
        token = generate_jwt(jwt_private, 12345)
        decoded = parse_jwt(token, jwt_public)
        assert decoded['id'] == 12345


class TestAuth:
    @staticmethod
    def test_signup(main_service_addr):
        login = str(uuid.uuid4())
        password = str(uuid.uuid4())
        r = make_requests(
            'POST',
            main_service_addr,
            '/signup',
            data={
                'login': login,
                'password': password})
        assert r.status_code == 200

    @staticmethod
    def test_signup_with_existing_user(main_service_addr, user):
        ((login, _), _) = user
        password = str(uuid.uuid4())
        r = make_requests(
            'POST',
            main_service_addr,
            '/signup',
            data={
                'login': login,
                'password': password})
        assert r.status_code == 403
        assert len(r.cookies) == 0

    @staticmethod
    def test_login(main_service_addr, user):
        ((login, password), _) = user
        r = make_requests(
            'POST',
            main_service_addr,
            '/login',
            data={
                'login': login,
                'password': password})
        assert r.status_code == 200

    @staticmethod
    def test_login_with_wrong_password(main_service_addr, user):
        ((login, _), _) = user
        password = str(uuid.uuid4())
        r = make_requests(
            'POST',
            main_service_addr,
            '/login',
            data={
                'login': login,
                'password': password})
        assert r.status_code == 403
        assert len(r.cookies) == 0

    @staticmethod
    def test_login_with_non_existing_user(main_service_addr):
        login = str(uuid.uuid4())
        password = str(uuid.uuid4())
        r = make_requests(
            'POST',
            main_service_addr,
            '/login',
            data={
                'login': login,
                'password': password})
        assert r.status_code == 403
        assert len(r.cookies) == 0

class TestUser:
    @staticmethod
    def test_profile(main_service_addr, user):
        ((_, _), cookies) = user
        r = make_requests(
            'POST',
            main_service_addr,
            '/profile',
            data={
                'firstName': str(uuid.uuid4()),
                'lastName': str(uuid.uuid4()),
                'birthdate': "2020-10-30",
                'email': str(uuid.uuid4()),
                'phoneNumber': str(uuid.uuid4())},
            cookies=cookies)
        assert r.status_code == 200

    @staticmethod
    def test_profile_with_wrong_cookie(main_service_addr):
        r = make_requests(
            'POST',
            main_service_addr,
            '/profile',
            data={
                'firstName': str(uuid.uuid4()),
                'lastName': str(uuid.uuid4()),
                'birthdate': "2020-10-30",
                'email': str(uuid.uuid4()),
                'phoneNumber': str(uuid.uuid4())},
            cookies={'jwt': 'not jwt'})
        assert r.status_code == 400