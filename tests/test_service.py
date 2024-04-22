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

LOGIN_HANDLER = "/user/login"
SIGNUP_HANDLER = "/user/signup"
PROFILE_HANDLER = "/user/profile"
POST_HANDLER = "/post"
POST_PAGE_HANDLER = "/post/page"

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
        SIGNUP_HANDLER,
        data={
            'login': login,
            'password': password})
    assert r.status_code == 200
    cookies = r.cookies.get_dict()
    return ((login, password), cookies)


def make_post(main_service_addr, user):
    ((_, _), cookies) = user
    r = make_requests(
        'POST',
        main_service_addr,
        POST_HANDLER,
        data={
            'content': str(uuid.uuid4()),
        },
        cookies=cookies
    )
    assert r.status_code == 200
    return r.json()


@pytest.fixture
def user(main_service_addr):
    yield make_user(main_service_addr)


@pytest.fixture
def user_with_post(main_service_addr):
    user = make_user(main_service_addr)
    yield user, make_post(main_service_addr, user)


def generate_jwt(private, id):
    return jwt.encode({'userID': id}, private, 'RS256')


def parse_jwt(token, public):
    return jwt.decode(token, public, ['RS256'])


def generate_hs256_jwt(secret, id):
    return jwt.encode({'userID': id}, secret, 'HS256')


class TestRSA:
    @staticmethod
    def test_private(jwt_private):
        generate_jwt(jwt_private, 12345)

    @staticmethod
    def test_public(jwt_private, jwt_public):
        token = generate_jwt(jwt_private, 12345)
        decoded = parse_jwt(token, jwt_public)
        assert decoded['userID'] == 12345


class TestAuth:
    @staticmethod
    def test_signup(main_service_addr):
        login = str(uuid.uuid4())
        password = str(uuid.uuid4())
        r = make_requests(
            'POST',
            main_service_addr,
            SIGNUP_HANDLER,
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
            SIGNUP_HANDLER,
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
            LOGIN_HANDLER,
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
            LOGIN_HANDLER,
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
            LOGIN_HANDLER,
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
            'PUT',
            main_service_addr,
            PROFILE_HANDLER,
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
            'PUT',
            main_service_addr,
            PROFILE_HANDLER,
            data={
                'firstName': str(uuid.uuid4()),
                'lastName': str(uuid.uuid4()),
                'birthdate': "2020-10-30",
                'email': str(uuid.uuid4()),
                'phoneNumber': str(uuid.uuid4())},
            cookies={'jwt': 'not jwt'})
        assert r.status_code == 401

class TestPost:
    @staticmethod
    def test_create_post(main_service_addr, user):
        make_post(main_service_addr, user)

    @staticmethod
    def test_get_post(main_service_addr, user_with_post):
        (user, post) = user_with_post
        ((_, _), cookies) = user
        new_content = str(uuid.uuid4())
        r = make_requests(
            'GET',
            main_service_addr,
            POST_HANDLER + "/" + str(post['postID']),
            data={
                'content': new_content,
            },
            cookies=cookies
        )
        assert r.status_code == 200
        assert r.json() == post

    @staticmethod
    def test_update_post(main_service_addr, user_with_post):
        (user, post) = user_with_post
        ((_, _), cookies) = user
        new_content = str(uuid.uuid4())
        r = make_requests(
            'PUT',
            main_service_addr,
            POST_HANDLER + "/" + str(post['postID']),
            data={
                'content': new_content,
            },
            cookies=cookies
        )
        assert r.status_code == 200
        assert r.json()['postID'] == post['postID']
        assert r.json()['userID'] == post['userID']
        assert r.json()['content'] == new_content
        assert r.json()['createdAt'] == post['createdAt']

    @staticmethod
    def test_delete_post(main_service_addr, user_with_post):
        (user, post) = user_with_post
        ((_, _), cookies) = user
        new_content = str(uuid.uuid4())
        r = make_requests(
            'DELETE',
            main_service_addr,
            POST_HANDLER + "/" + str(post['postID']),
            data={
                'content': new_content,
            },
            cookies=cookies
        )
        assert r.status_code == 200

    @staticmethod
    def test_get_page(main_service_addr, user):
        posts = []
        for _ in range(6):
            post = make_post(main_service_addr, user)
            posts.append(post)

        ((_, _), cookies) = user
        r = make_requests(
            'GET',
            main_service_addr,
            POST_PAGE_HANDLER,
            params={
                'page': 0,
                'pageSize': 5,
            },
            cookies=cookies
        )
        assert r.status_code == 200
        assert len(r.json()['posts']) == 5
        received = r.json()['posts']

        r = make_requests(
            'GET',
            main_service_addr,
            POST_PAGE_HANDLER,
            params={
                'page': 1,
                'pageSize': 5,
            },
            cookies=cookies
        )
        assert r.status_code == 200
        assert len(r.json()['posts']) == 1
        received += r.json()['posts']
        for i, post in enumerate(reversed(received)):
            assert post == posts[i]

    @staticmethod
    def test_no_access(main_service_addr, user_with_post):
        (_, post) = user_with_post
        ((_, _), cookies) = make_user(main_service_addr)
        new_content = str(uuid.uuid4())
        r = make_requests(
            'DELETE',
            main_service_addr,
            POST_HANDLER + "/" + str(post['postID']),
            data={
                'content': new_content,
            },
            cookies=cookies
        )
        assert r.status_code == 403
